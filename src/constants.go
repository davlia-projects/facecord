package main

// Numeric values
const (
	NumThreads = 32
)

// String values
const (
	AdminChannelName = "admin"
	LoginText        = "Please login to continue"
	UsernameText     = "Enter your username:"
	PasswordText     = "Enter your password:"
	LoginSuccessText = "Login successful!"
	LoginFailedText  = "Login failed, try again!"
)

// State
type AdminState int

const (
	Ready AdminState = iota + 1
	Executing
	NotAvailable
)

type Signal int

const Unblock Signal = 1
