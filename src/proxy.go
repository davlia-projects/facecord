package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/davlia/fbmsgr"
)

type FacebookProxy struct {
	dc      *discordgo.Session
	fb      *fbmsgr.Session
	guildID string
	inbox   chan *Message
	outbox  chan *Message
	Cache   *Cache
}

func NewFacebookProxy() (*FacebookProxy, error) {
	fb, err := fbmsgr.Auth(os.Getenv("FB_USERNAME"), os.Getenv("FB_PASSWORD"))
	if err != nil {
		panic("error authenticating. check your environment variables.")
	}

	dg, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(fmt.Sprintf("error creating Discord session,", err))
	}

	proxy := &FacebookProxy{
		dc:      dg,
		fb:      fb,
		guildID: os.Getenv("GUILD_ID"),
		inbox:   make(chan *Message),
		outbox:  make(chan *Message),
		Cache:   NewCache(),
	}
	return proxy, nil
}

func (T *FacebookProxy) Run() error {
	go T.runDiscordBot()
	go T.runFacebookClient()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	return nil
}

func (T *FacebookProxy) Stop() error {
	err := T.dc.Close()
	if err != nil {
		log.Printf("could not close discord session")
		return err
	}

	err = T.fb.Close()
	if err != nil {
		log.Printf("could not close facebook session")
		return err
	}

	return nil
}
