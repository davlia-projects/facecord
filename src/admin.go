package main

import (
	"fmt"
	"strings"

	"github.com/agext/levenshtein"
	"github.com/bwmarrin/discordgo"

	"github.com/davlia/fbmsgr"
	"github.com/facecord/src/logger"
)

func (T *ProxySession) adminPrintf(msg string, args ...interface{}) {
	m := fmt.Sprintf(msg, args...)
	T.dc.ChannelMessageSend(T.adminChannelID, m)
}

func (T *ProxySession) createAdminChannel() {
	channel, err := T.dc.GuildChannelCreate(T.guildID, AdminChannelName, "text")
	T.registerChannel(channel)
	if err != nil {
		logger.Error(NoTag, "could not create admin channel: %s\n", err)
	}
	T.adminChannelID = channel.ID
	T.adminPrintf(LoginText)
}

func (T *ProxySession) createChannel(name string) (string, error) {
	channel, err := T.dc.GuildChannelCreate(T.guildID, name, "text")
	if err != nil {
		return "", err
	}
	T.registerChannel(channel)
	return channel.ID, nil
}

func (T *ProxySession) purgeChannels() {
	channels, err := T.dc.GuildChannels(T.guildID)
	if err != nil {
		logger.Error(NoTag, "could not purge channels: %s\n", err)
		return
	}
	for _, ch := range channels {
		if ch.ID != T.adminChannelID {
			T.deleteChannel(ch.ID)
		}
	}
}

/**
 * Admin commands
 */

func (T *ProxySession) cmdHelp() {
	T.adminPrintf(HelpText)
}

func (T *ProxySession) cmdLogin(args []string) {
	var err error
	if len(args) != 2 || T.fb != nil {
		return
	}
	T.fb, err = fbmsgr.Auth(args[0], args[1])
	if err != nil {
		logger.Error(NoTag, "error authenticating")
		T.adminPrintf(LoginFailedText)
		return
	}
	T.adminPrintf(LoginSuccessText)
	T.updateFriends()
	entries := T.updateThreads(NumThreads)
	T.renderEntries(entries)
	T.fbInbox = make(chan *Message)
	T.fbOutbox = make(chan *Message)
	go T.runFacebookClient()
	go T.consumeFbInbox()
	go T.handleOutboundMessage()
}

func (T *ProxySession) cmdLogout() {
	T.fb.Close()
	close(T.fbInbox)
	close(T.fbOutbox)
	T.purgeChannels()
	T.fb = nil
}

func (T *ProxySession) cmdOpen(args []string) {
	var (
		mostSimilar  *Entry
		highestScore float64
	)
	name := NormalizeStr(strings.Join(args, " "))
	p := levenshtein.NewParams()
	for entry := range T.cache.Entries {
		score := levenshtein.Similarity(name, NormalizeStr(entry.Name), p)
		if score > highestScore {
			highestScore = score
			mostSimilar = entry
		}
	}
	if mostSimilar.ChannelID == "" {
		channelID, err := T.createChannel(mostSimilar.Name)
		if err != nil {
			logger.Error(NoTag, "error handling cmdOpen: %+v\n", err)
			return
		}
		mostSimilar.ChannelID = channelID
	}
}

func (T *ProxySession) cmdClose(args []string) {
	var (
		mostSimilar  *discordgo.Channel
		highestScore float64
	)
	name := FmtDiscordChannelName(strings.Join(args, " "))
	p := levenshtein.NewParams()
	channels, err := T.dc.GuildChannels(T.guildID)
	if err != nil {
		logger.Error(NoTag, "error grabbing list of channels: %+v\n", err)
		return
	}
	for _, ch := range channels {
		score := levenshtein.Similarity(name, ch.Name, p)
		if score > highestScore {
			highestScore = score
			mostSimilar = ch
		}
	}
	T.deleteChannel(mostSimilar.ID)
}

func (T *ProxySession) cmdCloseAll() {
	T.purgeChannels()
}
