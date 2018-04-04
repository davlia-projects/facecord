package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func (T *ProxyBot) runDiscordBot() error {
	T.dc.AddHandler(T.guildCreate)
	T.dc.AddHandler(T.routeMessages)

	err := T.dc.Open()
	if err != nil {
		log.Printf("error: %s\n", err)
		return err
	}

	return nil
}

func (T *ProxyBot) guildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	session := NewProxySession(e.Guild.ID, s, T.registry)
	session.Setup()
}

func (T *ProxyBot) routeMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	c, err := T.registry.Lookup(m.ChannelID)
	if err != nil {
		return
	}
	*c <- m.Message
}
