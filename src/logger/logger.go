package logger

import (
	"fmt"
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

func printf(level Level, tag, msg string, args ...interface{}) {
	if l.level <= level {
		tagFmt := msg
		if tag != "" {
			tagFmt = fmt.Sprintf("%s: %s\n", tag, msg)
		}
		log.Printf(tagFmt, args...)
	}
}

func Info(tag, msg string, args ...interface{}) {
	printf(InfoLevel, tag, msg, args...)
}

func Debug(tag, msg string, args ...interface{}) {
	printf(DebugLevel, tag, msg, args...)
}

func Error(tag, msg string, args ...interface{}) {
	printf(ErrorLevel, tag, msg, args...)
}
