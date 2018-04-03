package main

type Message struct {
	FBID  string
	Name  string
	Body  string
	Group string
}

func NewMessage(fbid, name, body, group string) *Message {
	m := &Message{
		FBID:  fbid,
		Name:  name,
		Body:  body,
		Group: group,
	}
	return m
}
