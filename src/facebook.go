package main

import (
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
	log.Printf("received message: %+v\n", msg)
	T.inbox <- NewMessage(msg.SenderFBID, msg.OtherUser, msg.Body, msg.GroupThread)

}

func (T *FacebookProxy) handleOutboundMessage() {
	for {
		select {
		case msg := <-T.outbox:
			T.fb.SendText(msg.ID, msg.Body)
		}
	}
}

func (T *FacebookProxy) fetchThread(threadID string) *fbmsgr.ThreadInfo {
	thread, err := T.fb.Thread(threadID)
	if err != nil {
		log.Printf("could not find thread: %+v\n", err)
		return nil
	}
	return thread
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
func (T *FacebookProxy) fetchFriend(fbid string) *fbmsgr.FriendInfo {
	friend, err := T.fb.Friend(fbid)
	if err != nil {
		log.Printf("could not find friend: %+v\n", err)
		return nil
	}
	return friend
}

func (T *FacebookProxy) fetchFriends() map[string]*fbmsgr.FriendInfo {
	friends, err := T.fb.Friends()
	if err != nil {
		panic(err)
	}
	return friends
}
