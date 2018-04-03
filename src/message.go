package main

type Message struct {
	ID    string
	Name  string
	Body  string
	Group string
}

func NewMessage(id, name, body, group string) *Message {
	m := &Message{
		ID:    id,
		Name:  name,
		Body:  body,
		Group: group,
	}
	return m
}
