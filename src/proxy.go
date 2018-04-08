package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/facecord/src/logger"
)

type ProxyBot struct {
	dc       *discordgo.Session
	inbox    chan *Message
	outbox   chan *Message
	registry *Registry
}

func NewProxyBot() (*ProxyBot, error) {
	dg, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(fmt.Sprintf("error creating Discord session,", err))
	}

	proxy := &ProxyBot{
		dc:       dg,
		inbox:    make(chan *Message),
		outbox:   make(chan *Message),
		registry: NewRegistry(),
	}
	return proxy, nil
}

func (T *ProxyBot) Run() error {
	go T.runDiscordBot()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	return nil
}

func (T *ProxyBot) Stop() error {
	err := T.dc.Close()
	if err != nil {
		logger.Error(NoTag, "could not close discord session")
		return err
	}

	return nil
}
