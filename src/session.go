package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/facecord/src/logger"

	"github.com/davlia/fbmsgr"
)

type ProxySession struct {
	guildID        string
	adminChannelID string
	fbInbox        chan *Message
	fbOutbox       chan *Message
	dcInbox        chan *discordgo.Message
	cache          *Cache
	fb             *fbmsgr.Session
	dc             *discordgo.Session
	registry       *Registry

	// this is some hacky shit idek if this is idiomatic go
	AdminState   AdminState
	block        chan interface{}
	results      chan interface{}
	AdminHandler *func(ps *ProxySession, m *discordgo.Message)
}

func NewProxySession(guildID string, dc *discordgo.Session, registry *Registry) *ProxySession {
	r := ready
	ps := &ProxySession{
		guildID:      guildID,
		fbInbox:      make(chan *Message),
		fbOutbox:     make(chan *Message),
		dcInbox:      make(chan *discordgo.Message),
		cache:        NewCache(),
		dc:           dc,
		registry:     registry,
		AdminState:   Ready,
		block:        make(chan interface{}, 1),
		results:      make(chan interface{}),
		AdminHandler: &r,
	}
	return ps
}

func (T *ProxySession) Run() {
	go T.consumeDcInbox()

	T.purgeChannels()
	T.createAdminChannel()
	T.authenticate()

}

func (T *ProxySession) registerChannel(channel *discordgo.Channel) {
	T.registry.Register(channel.ID, &T.dcInbox)
}

func (T *ProxySession) renderEntries(entries []*Entry) {
	for _, entry := range entries {
		if entry.ChannelID == "" && entry.Name != "" {
			channelID, err := T.createChannel(entry.Name)
			if err != nil {
				logger.Error(NoTag, "error creating channel: %s\n", err)
				continue
			}
			entry.ChannelID = channelID
			T.cache.upsertEntry(entry)
		}
	}
}

func (T *ProxySession) createChannel(name string) (string, error) {
	channel, err := T.dc.GuildChannelCreate(T.guildID, name, "text")
	if err != nil {
		return "", err
	}
	T.registerChannel(channel)
	return channel.ID, nil
}

/**
 * Handle incoming messages from messenger API
 */

func (T *ProxySession) consumeFbInbox() {
	for {
		select {
		case msg := <-T.fbInbox:
			T.handleInboxMessage(msg)
		}
	}
}

func (T *ProxySession) handleInboxMessage(msg *Message) {
	if msg.Group == "" {
		T.handleDirectMessage(msg)
	} else {
		T.handleGroupMessage(msg)
	}
}

func (T *ProxySession) handleGroupMessage(msg *Message) {
	fbid := msg.Group
	entry, err := T.cache.getByFBID(fbid)
	if err != nil {
		// Fetch and cache
		thread := T.fetchThread(fbid)
		entry = &Entry{
			FBID:    fbid,
			IsGroup: true,
		}
		if thread.Name != "" {
			entry.Name = thread.Name
		} else {
			entry.Name = fbid
		}
		entry.ChannelID, err = T.createChannel(entry.Name)
		if err != nil {
			logger.Error(NoTag, "error while handling facebook inbox message: %s\n", err)
			return
		}
		T.cache.upsertEntry(entry)
	}
	// Get the sender name
	var senderName string
	sender, err := T.cache.getByFBID(msg.FBID)
	if err != nil {
		friend := T.fetchFriend(msg.FBID)
		senderName = friend.AlternateName
	} else {
		senderName = sender.Name
	}
	embed := CreateMessageEmbed(senderName, msg.Body)
	T.dc.ChannelMessageSendEmbed(entry.ChannelID, embed)
}

func (T *ProxySession) handleDirectMessage(msg *Message) {
	fbid := msg.FBID
	entry, err := T.cache.getByFBID(fbid)
	if err != nil {
		// Fetch and cache
		friend := T.fetchFriend(fbid)
		entry = &Entry{
			Name: friend.Vanity,
			FBID: fbid,
		}
		entry.ChannelID, err = T.createChannel(entry.Name)
		if err != nil {
			logger.Error(NoTag, "error while handling facebook inbox message: %s\n", err)
			return
		}
		T.cache.upsertEntry(entry)
	}
	embed := CreateMessageEmbed(entry.Name, msg.Body)
	T.dc.ChannelMessageSendEmbed(entry.ChannelID, embed)
}

/**
 * Handle incoming messages from Discord
 */

func (T *ProxySession) consumeDcInbox() {
	for {
		select {
		case msg := <-T.dcInbox:
			T.handleDiscordMessage(msg)
		}
	}
}

func (T *ProxySession) handleDiscordMessage(m *discordgo.Message) {
	if m.ChannelID == T.adminChannelID {
		(*T.AdminHandler)(T, m)
	} else {
		T.forwardFbMessage(m)
	}
}

func (T *ProxySession) forwardFbMessage(m *discordgo.Message) {
	var msg *Message

	entry, err := T.cache.getByChannelID(m.ChannelID)
	if err != nil {
		logger.Error(NoTag, "error while forwarding messages: %s. entry: %s\n", err, entry)
		return
	}

	if entry.IsGroup {
		msg = &Message{
			Group: entry.FBID,
			Body:  m.Content,
		}
	} else {
		msg = &Message{
			FBID: entry.FBID,
			Body: m.Content,
		}
	}

	T.fbOutbox <- msg
}
