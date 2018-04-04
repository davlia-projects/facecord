package main

import (
	"log"

	"github.com/davlia/fbmsgr"
)

func (T *ProxySession) runFacebookClient() {
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

func (T *ProxySession) handleInboundMessage(msg fbmsgr.MessageEvent) {
	if msg.SenderFBID == T.fb.FBID() {
		return
	}
	log.Printf("received message: %+v\n", msg)
	T.fbInbox <- NewMessage(msg.SenderFBID, msg.OtherUser, msg.Body, msg.GroupThread)

}

func (T *ProxySession) handleOutboundMessage() {
	for {
		select {
		case msg := <-T.fbOutbox:
			if msg.Group == "" {
				T.fb.SendText(msg.FBID, msg.Body)
			} else {
				T.fb.SendGroupText(msg.Group, msg.Body)
			}
		}
	}
}

func (T *ProxySession) fetchThread(threadID string) *fbmsgr.ThreadInfo {
	thread, err := T.fb.Thread(threadID)
	if err != nil {
		log.Printf("could not find thread: %+v\n", err)
		return nil
	}
	return thread
}

func (T *ProxySession) fetchThreads() []*fbmsgr.ThreadInfo {
	threads := []*fbmsgr.ThreadInfo{}
	idx := 0
	result, err := T.fb.Threads(idx, BatchSize)
	if err != nil {
		panic(err)
	}
	threads = append(threads, result.Threads...)
	return threads
}
func (T *ProxySession) fetchFriend(fbid string) *fbmsgr.FriendInfo {
	friend, err := T.fb.Friend(fbid)
	if err != nil {
		log.Printf("could not find friend: %+v\n", err)
		return nil
	}
	return friend
}

func (T *ProxySession) fetchFriends() map[string]*fbmsgr.FriendInfo {
	friends, err := T.fb.Friends()
	if err != nil {
		panic(err)
	}
	return friends
}
