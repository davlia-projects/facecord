package main

import (
	"errors"
)

// should i set up redis :thinking:
type Cache struct {
	FBIDMap      map[string]*Entry
	NameMap      map[string]*Entry
	ChannelIDMap map[string]*Entry
	Entries      map[*Entry]bool
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
		Entries:      make(map[*Entry]bool),
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
	if entry.FBID != "" {
		T.FBIDMap[entry.FBID] = entry
	}
	if entry.Name != "" {
		T.NameMap[entry.Name] = entry
	}
	if entry.ChannelID != "" {
		T.ChannelIDMap[entry.ChannelID] = entry
	}
	T.Entries[entry] = true
}

func (T *Cache) deleteEntry(entry *Entry) {
	// TODO: implement
}

func (T *ProxySession) updateThreads(numThreads int) []*Entry {
	entries := []*Entry{}
	threads := T.fetchThreads(numThreads)
	for _, thread := range threads {
		entry := &Entry{
			Name: thread.Name,
		}
		if thread.OtherUserFBID != nil && *thread.OtherUserFBID != "" {
			entry.FBID = *thread.OtherUserFBID
		} else {
			entry.FBID = thread.ThreadFBID
			entry.IsGroup = true
			// TODO: let's give it a better name...
			if entry.Name == "" {
				entry.Name = entry.FBID
			}
		}
		T.cache.upsertEntry(entry)
		entries = append(entries, entry)
	}
	return entries
}

func (T *ProxySession) updateFriends() []*Entry {
	entries := []*Entry{}
	friends := T.fetchFriends()
	for fbid, friend := range friends {
		entry := &Entry{
			FBID: fbid,
			Name: friend.AlternateName,
		}
		T.cache.upsertEntry(entry)
		entries = append(entries, entry)
	}
	return entries
}
