package main

import (
	"math"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/facecord/src/logger"

	"github.com/davlia/fbmsgr"
)

type ProxySession struct {
	guildID        string
	adminChannelID string
	dmCategoryID   string
	firstPosition  int
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
		guildID:       guildID,
		firstPosition: math.MaxInt32 - 1,
		dcInbox:       make(chan *discordgo.Message),
		cache:         NewCache(),
		dc:            dc,
		registry:      registry,
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

func (T *ProxySession) unregisterChannel(channel *discordgo.Channel) {
	T.registry.Unregister(channel.ID)
}

func (T *ProxySession) renderEntries(entries []*Entry) {
	for _, entry := range entries {
		if entry.ChannelID == "" && entry.Name != "" {
			channelID, err := T.createConversation(entry.Name)
			if err != nil {
				logger.Error(NoTag, "error creating channel: %s\n", err)
				continue
			}
			entry.ChannelID = channelID
			T.cache.upsertEntry(entry)
		}
	}
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
		entry.ChannelID, err = T.createConversation(entry.Name)
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
	T.moveChannelToTop(entry.ChannelID)
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
		entry.ChannelID, err = T.createConversation(entry.Name)
		if err != nil {
			logger.Error(NoTag, "error while handling facebook inbox message: %s\n", err)
			return
		}
		T.cache.upsertEntry(entry)
	}
	embed := CreateMessageEmbed(entry.Name, msg.Body)
	T.moveChannelToTop(entry.ChannelID)
	T.dc.ChannelMessageSendEmbed(entry.ChannelID, embed)
}

func (T *ProxySession) moveChannelToTop(channelID string) {
	ch, err := T.dc.Channel(channelID)
	if err != nil {
		logger.Error(NoTag, "error while retrieving channel %s: %s\n", channelID, err)
		return
	}
	if ch.Position != T.firstPosition {
		T.dc.ChannelEditComplex(channelID, &discordgo.ChannelEdit{
			Position: T.firstPosition,
		})
		T.firstPosition--
	}
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
	switch args := toks[1:]; toks[0] {
	case "!help":
		T.cmdHelp()
	case "!login":
		T.dc.ChannelMessageDelete(T.adminChannelID, m.ID)
		T.cmdLogin(args)
	case "!logout":
		T.cmdLogout()
	case "!open":
		T.cmdOpen(args)
	case "!close":
		T.cmdClose(args)
	case "!close-all":
		T.cmdCloseAll()
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

	T.moveChannelToTop(m.ChannelID)
	T.fbOutbox <- msg
}
