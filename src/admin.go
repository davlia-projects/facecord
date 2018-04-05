package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/davlia/fbmsgr"
)

var (
	r = ready
)

func (T *ProxySession) unblock(unblock interface{}) {
	select {
	case T.block <- unblock:
		log.Printf("unblocking with: %s\n", unblock)
	default:
		log.Printf("tossing away unblock %s\n", unblock)
	}
}

func (T *ProxySession) expect() {
	log.Printf("blocking with expect\n")
	debug := <-T.block
	log.Printf("got what i expected: %s\n", debug)
	// T.results <- debug
	log.Printf("unblocking expectInput and proceeding with next step\n")
}

func ready(ps *ProxySession, m *discordgo.Message) {
	if m.Author.ID == ps.dc.State.User.ID {
		return
	}
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
	T.expect()
	T.AdminHandler = &r
	return T
}

func (T *ProxySession) login() {
	var err error
	T.fb, err = fbmsgr.Auth((<-T.results).(string), (<-T.results).(string))
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
	if m.Author.ID == ps.dc.State.User.ID {
		return
	}
	msg := m.Content
	toks := strings.Split(msg, " ")
	if len(toks) == 1 {
		ps.unblock(toks[0])
	}
}

func password(ps *ProxySession, m *discordgo.Message) {
	if m.Author.ID == ps.dc.State.User.ID {
		return
	}
	msg := m.Content
	toks := strings.Split(msg, " ")
	if len(toks) == 1 {
		ps.unblock(toks[0])
	}
}
