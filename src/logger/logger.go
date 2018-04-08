package logger

import (
	"log"
)

type Level int

const (
	DebugLevel = iota
	InfoLevel
	ErrorLevel
)

var l = Logger{}

type Logger struct {
	level Level
}

func SetLevel(level Level) {
	l.level = level
}

func Info(fmt string, args ...interface{}) {
	if l.level < InfoLevel {
		log.Printf(fmt, args...)
	}
}
