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
	T.updateFBIDs()
	// T.syncGuildChannels()

	go T.consumeInbox()

	return nil
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
		T.store.upsertEntry(entry)
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
	entry, err := T.store.getByFBID(msg.ID)
	if err != nil {
		return
	}
	if entry.ChannelID == "" {
		entry.ChannelID, err = T.createChannel(entry.Name)
		if err != nil {
			return
		}
	}

	T.dc.ChannelMessageSend(entry.ChannelID, msg.Body)
}

func (T *FacebookProxy) forwardMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	entry, err := T.store.getByFBID(m.ChannelID)
	if err != nil {
		log.Printf("error: %s\n", err)
		return
	}
	msg := &Message{
		ID:   entry.FBID,
		Body: m.Content,
	}

	T.outbox <- msg

}
