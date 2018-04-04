package main

import (
	"errors"
)

// should i set up redis :thinking:
type Cache struct {
	FBIDMap      map[string]*Entry
	NameMap      map[string]*Entry
	ChannelIDMap map[string]*Entry
}

type Entry struct {
	FBID      string
	Name      string
	ChannelID string
	IsGroup   bool
}

func NewCache() *Cache {
	s := &Cache{
		FBIDMap:      make(map[string]*Entry),
		NameMap:      make(map[string]*Entry),
		ChannelIDMap: make(map[string]*Entry),
	}
	return s
}

func (T *Cache) getByFBID(fbid string) (*Entry, error) {
	if entry, ok := T.FBIDMap[fbid]; ok {
		return entry, nil
	}
	return nil, errors.New("entry not found")
}

func (T *Cache) getByName(name string) (*Entry, error) {
	if entry, ok := T.NameMap[name]; ok {
		return entry, nil
	}
	return nil, errors.New("entry not found")
}

func (T *Cache) getByChannelID(channelID string) (*Entry, error) {
	if entry, ok := T.ChannelIDMap[channelID]; ok {
		return entry, nil
	}
	return nil, errors.New("entry not found")
}

func (T *Cache) upsertEntry(entry *Entry) {
	T.FBIDMap[entry.FBID] = entry
	T.NameMap[entry.Name] = entry
	T.ChannelIDMap[entry.ChannelID] = entry
}

func (T *Cache) deleteEntry(entry *Entry) {
	// TODO: implement
}

func (T *ProxySession) updateFBIDs() {
	threads := T.fetchThreads()
	for _, thread := range threads {
		entry := &Entry{
			Name: thread.Name,
		}
		if thread.OtherUserFBID != nil && *thread.OtherUserFBID != "" {
			entry.FBID = *thread.OtherUserFBID
		} else {
			entry.FBID = thread.ThreadFBID
		}
		T.Cache.upsertEntry(entry)
	}
}

func (T *ProxySession) populateCache() {
	friends := T.fetchFriends()
	for fbid, friend := range friends {
		entry := &Entry{
			FBID: fbid,
			Name: friend.FullName,
		}
		T.Cache.upsertEntry(entry)
	}
}
