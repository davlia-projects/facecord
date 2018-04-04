package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/davlia/fbmsgr"
)

func (T *ProxySession) unblock(unblock interface{}) {
	select {
	case T.Block <- unblock:
	default:
	}
}

func (T *ProxySession) block() interface{} {
	return <-T.Block
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
func (T *ProxySession) authenticate() {
	T.
		prompt(LoginText).
		prompt(UsernameText).
		expectInput(username).
		prompt(PasswordText).
		expectInput(password).
		login()
}

func (T *ProxySession) createAdminChannel() {
	channel, err := T.dc.GuildChannelCreate(T.guildID, AdminChannelName, "text")
	T.registerChannel(channel)
	if err != nil {
		log.Printf("could not create admin channel: %s\n", err)
	}
	T.adminChannelID = channel.ID
}

func (T *ProxySession) prompt(text string) *ProxySession {
	T.dc.ChannelMessageSend(T.adminChannelID, text)
	return T
}

func (T *ProxySession) expectInput(handler func(ps *ProxySession, m *discordgo.Message)) *ProxySession {
	T.AdminHandler = &handler
	return T
}

func (T *ProxySession) login() {
	var err error
	T.fb, err = fbmsgr.Auth(T.block().(string), T.block().(string))
	if err != nil {
		log.Printf("error authenticating")
		T.dc.ChannelMessageSend(T.adminChannelID, LoginFailedText)
		return
	}
	T.dc.ChannelMessageSend(T.adminChannelID, LoginSuccessText)
	T.updateFriends()
	entries := T.updateThreads(NumThreads)
	T.renderEntries(entries)
	go T.runFacebookClient()
	go T.consumeFbInbox()
}

func username(ps *ProxySession, m *discordgo.Message) {
	msg := m.Content
	toks := strings.Split(msg, " ")
	if len(toks) == 1 {
		ps.unblock(toks[0])
	}
}

func password(ps *ProxySession, m *discordgo.Message) {
	msg := m.Content
	toks := strings.Split(msg, " ")
	if len(toks) == 1 {
		ps.unblock(toks[0])
	}
}
