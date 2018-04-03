package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func (T *FacebookProxy) runDiscordBot() error {
	T.dc.AddHandler(T.forwardMessage)
	err := T.dc.Open()
	if err != nil {
		log.Printf("error: %s\n", err)
		return err
	}
	T.purgeChannels() //debug
	// T.populateCache()
	// T.updateFBIDs()
	// T.syncGuildChannels()

	go T.consumeInbox()

	return nil
}

func (T *FacebookProxy) populateCache() {
	friends := T.fetchFriends()
	for fbid, friend := range friends {
		entry := &Entry{
			FBID: fbid,
			Name: friend.FullName,
		}
		T.Cache.upsertEntry(entry)
	}
}

func (T *FacebookProxy) purgeChannels() {
	channels, err := T.dc.GuildChannels(T.guildID)
	if err != nil {
		log.Printf("error: %s\n", err)
		return
	}
	for _, ch := range channels {
		T.dc.ChannelDelete(ch.ID)
	}
}

func (T *FacebookProxy) updateFBIDs() {
	threads := T.fetchThreads()
	for _, thread := range threads {
		entry := &Entry{
			Name: thread.Name,
		}
		if thread.OtherUserFBID != nil && *thread.OtherUserFBID != "" {
			entry.FBID = *thread.OtherUserFBID
		} else {
			entry.FBID = thread.ThreadFBID
		}
		T.Cache.upsertEntry(entry)
	}
}

func (T *FacebookProxy) syncGuildChannels() {
	// TODO: is this necessary?
}

func (T *FacebookProxy) createChannel(name string) (string, error) {
	channel, err := T.dc.GuildChannelCreate(T.guildID, name, "text")
	if err != nil {
		log.Printf("error: %s\n", err)
		return "", nil
	}
	return channel.ID, nil
}

func (T *FacebookProxy) consumeInbox() {
	for {
		select {
		case msg := <-T.inbox:
			T.handleInboxMessage(msg)
		}
	}
}

func (T *FacebookProxy) handleInboxMessage(msg *Message) {
	if msg.Group == "" {
		T.handleDirectMessage(msg)
	} else {
		T.handleGroupMessage(msg)
	}
}

func (T *FacebookProxy) handleGroupMessage(msg *Message) {
	fbid := msg.Group
	entry, err := T.Cache.getByFBID(fbid)
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
			log.Printf("error while handling inbox message: %s\n", err)
			return
		}
		T.Cache.upsertEntry(entry)
	}
	// Get the sender name
	var senderName string
	sender, err := T.Cache.getByFBID(msg.FBID)
	if err != nil {
		friend := T.fetchFriend(msg.FBID)
		senderName = friend.Vanity
	} else {
		senderName = sender.Name
	}
	embed := CreateMessageEmbed(senderName, msg.Body)
	T.dc.ChannelMessageSendEmbed(entry.ChannelID, embed)
}

func (T *FacebookProxy) handleDirectMessage(msg *Message) {
	fbid := msg.FBID
	entry, err := T.Cache.getByFBID(fbid)
	if err != nil {
		// Fetch and cache
		friend := T.fetchFriend(fbid)
		entry = &Entry{
			Name: friend.Vanity,
			FBID: fbid,
		}
		entry.ChannelID, err = T.createChannel(entry.Name)
		if err != nil {
			log.Printf("error while handling inbox message: %s\n", err)
			return
		}
		T.Cache.upsertEntry(entry)
	}
	embed := CreateMessageEmbed(entry.Name, msg.Body)
	T.dc.ChannelMessageSendEmbed(entry.ChannelID, embed)
}

func (T *FacebookProxy) forwardMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	var msg *Message
	if m.Author.ID == s.State.User.ID {
		return
	}
	entry, err := T.Cache.getByChannelID(m.ChannelID)
	if err != nil {
		log.Printf("error while forwarding messages: %s\n", err)
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

	T.outbox <- msg

}
