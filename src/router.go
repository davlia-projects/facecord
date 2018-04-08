package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/facecord/src/logger"
)

func (T *ProxyBot) runDiscordBot() error {
	T.dc.AddHandler(T.guildCreate)
	T.dc.AddHandler(T.routeMessages)

	err := T.dc.Open()
	if err != nil {
		logger.Error(NoTag, "could not create discord bot session: %s\n", err)
		return err
	}

	return nil
}

func (T *ProxyBot) guildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	logger.Info(NoTag, "Creating new session")
	session := NewProxySession(e.Guild.ID, s, T.registry)
	go session.Run()
}

func (T *ProxyBot) routeMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	c, err := T.registry.Lookup(m.ChannelID)
	if err != nil {
		return
	}
	*c <- m.Message
}
