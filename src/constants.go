package main

// Numeric values
const (
	NumThreads = 32
)

// String values
const (
	NoTag            = ""
	Unhandled        = "Unhandled"
	Received         = "Message Received"
	AdminChannelName = "admin"
	LoginText        = "Please login to continue!\nType `!login <username> <password>`"
	LoginSuccessText = "Login successful!"
	LoginFailedText  = "Login failed, try again!"
	HelpText         = "Commands:\n\n" +
		"`!login <username> <password>`:\n" +
		"Authenticates you using your Facebook username and password\n\n" +
		"`!open <first name> <last name>`:\n" +
		"Opens a chat with a friend matching first name and last name\n\n" +
		"`!close <first name> <last name>`:\n" +
		"Close a chat with a friend matching first name and last name\n\n" +
		"`!close-all`:\n" +
		"Close all chats with all of your friends\n\n" +
		"`!help`:\n" +
		"What you're currently reading right now :)"
)
