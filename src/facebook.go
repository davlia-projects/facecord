package main

import (
	"fmt"
	"log"

	"github.com/davlia/fbmsgr"
)

const (
	BatchSize  = 32
	NumThreads = 20
)

func (T *FacebookProxy) runFacebookClient() {
	stream := T.fb.EventStream()

	go T.handleOutboundMessage()

	defer stream.Close()
	for {
		select {
		case evt := <-stream.Chan():
			if msg, ok := evt.(fbmsgr.MessageEvent); ok {
				T.handleInboundMessage(msg)
			} else {
				log.Printf("unhandled event\n")
			}
		}
	}
}

func (T *FacebookProxy) handleInboundMessage(msg fbmsgr.MessageEvent) {
	if msg.SenderFBID == T.fb.FBID() {
		return
	}
	fmt.Println("received message:", msg)
	T.inbox <- NewMessage(msg.SenderFBID, "", msg.Body, msg.GroupThread)

}

func (T *FacebookProxy) handleOutboundMessage() {
	for {
		select {
		case msg := <-T.outbox:
			T.fb.SendText(msg.ID, msg.Body)
		}
	}
}

func (T *FacebookProxy) fetchThreads() []*fbmsgr.ThreadInfo {
	threads := []*fbmsgr.ThreadInfo{}
	idx := 0
	result, err := T.fb.Threads(idx, BatchSize)
	if err != nil {
		panic(err)
	}
	threads = append(threads, result.Threads...)
	return threads
}
