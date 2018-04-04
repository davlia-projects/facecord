package main

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

type Registry struct {
	registry map[string]*chan *discordgo.Message
}

func NewRegistry() *Registry {
	r := &Registry{
		registry: make(map[string]*chan *discordgo.Message),
	}
	return r
}

func (T *Registry) Register(channelID string, ch *chan *discordgo.Message) {
	T.registry[channelID] = ch
}

func (T *Registry) Lookup(channelID string) (*chan *discordgo.Message, error) {
	if ch, ok := T.registry[channelID]; ok {
		return ch, nil
	}
	return nil, errors.New("channel could not be found in registry")
}
