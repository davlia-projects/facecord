package main

import (
	"fmt"

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
			T.dc.ChannelDelete(ch.ID)
		}
	}
}

/**
 * Admin commands
 */

func (T *ProxySession) cmdHelp() {
	T.adminPrintf(HelpText)
}

func (T *ProxySession) cmdAuthenticate(args []string) {
	if len(args) != 2 {
		return
	}
	fb, err := fbmsgr.Auth(args[0], args[1])
	if err != nil {
		logger.Error(NoTag, "error authenticating")
		T.adminPrintf(LoginFailedText)
		return
	}
	T.adminPrintf(LoginSuccessText)
	T.fb = fb
	T.updateFriends()
	entries := T.updateThreads(NumThreads)
	T.renderEntries(entries)
	go T.runFacebookClient()
	go T.consumeFbInbox()
}

func (T *ProxySession) cmdOpen(args []string) {
	// TODO: Implement
}

func (T *ProxySession) cmdClose(args []string) {
	// TODO: Implement
}

func (T *ProxySession) cmdCloseAll() {
	T.purgeChannels()
}
