package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func FmtDiscordChannelName(name string) string {
	name = strings.ToLower(name)
	toks := strings.Split(name, " ")
	name = strings.Join(toks, "-")
	return name
}

func CreateMessageEmbed(name, body string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: name,
		},
		Description: body,
	}
	return embed
}
