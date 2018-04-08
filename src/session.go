package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

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
}

func NewProxySession(guildID string, dc *discordgo.Session, registry *Registry) *ProxySession {
	ps := &ProxySession{
		guildID:  guildID,
		fbInbox:  make(chan *Message),
		fbOutbox: make(chan *Message),
		dcInbox:  make(chan *discordgo.Message),
		cache:    NewCache(),
		dc:       dc,
		registry: registry,
	}
	return ps
}

func (T *ProxySession) Run() {
	T.purgeChannels()
	T.createAdminChannel()
	go T.consumeDcInbox()

}

func (T *ProxySession) registerChannel(channel *discordgo.Channel) {
	T.registry.Register(channel.ID, &T.dcInbox)
}

func (T *ProxySession) purgeChannels() {
	channels, err := T.dc.GuildChannels(T.guildID)
	if err != nil {
		log.Printf("error: %s\n", err)
		return
	}
	for _, ch := range channels {
		T.dc.ChannelDelete(ch.ID)
	}
}

func (T *ProxySession) createAdminChannel() {
	channel, err := T.dc.GuildChannelCreate(T.guildID, AdminChannelName, "text")
	T.registerChannel(channel)
	if err != nil {
		log.Printf("could not create admin channel: %s\n", err)
	}
	T.adminChannelID = channel.ID
	T.dc.ChannelMessageSend(T.adminChannelID, LoginText)
}

func (T *ProxySession) authenticate(username, password string) {
	fb, err := fbmsgr.Auth(username, password)
	if err != nil {
		log.Printf("error authenticating")
		T.dc.ChannelMessageSend(T.adminChannelID, LoginFailedText)
		return
	}
	T.dc.ChannelMessageSend(T.adminChannelID, LoginSuccessText)
	T.fb = fb
	T.updateFriends()
	entries := T.updateThreads(NumThreads)
	T.renderEntries(entries)
	go T.runFacebookClient()
	go T.consumeFbInbox()

}

func (T *ProxySession) renderEntries(entries []*Entry) {
	for _, entry := range entries {
		if entry.ChannelID == "" && entry.Name != "" {
			channelID, err := T.createChannel(entry.Name)
			if err != nil {
				log.Printf("error creating channel: %s\n", err)
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
			log.Printf("error while handling fbInbox message: %s\n", err)
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
			log.Printf("error while handling fbInbox message: %s\n", err)
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
		T.handleAdminMessage(m)
	} else {
		T.forwardFbMessage(m)
	}
}
func (T *ProxySession) handleAdminMessage(m *discordgo.Message) {
	msg := m.Content
	toks := strings.Split(msg, " ")
	if toks[0] == "!login" && len(toks) == 3 {
		T.authenticate(toks[1], toks[2])
	}
}

func (T *ProxySession) forwardFbMessage(m *discordgo.Message) {
	var msg *Message

	entry, err := T.cache.getByChannelID(m.ChannelID)
	if err != nil {
		log.Printf("error while forwarding messages: %s. entry: %s\n", err, entry)
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