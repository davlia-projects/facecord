package main

import (
	"strings"
)

func FmtDiscordChannelName(name string) string {
	name = strings.ToLower(name)
	toks := strings.Split(name, " ")
	name = strings.Join(toks, "-")
	return name
}
