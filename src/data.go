package main

import (
	"errors"
)

// too lazy to set up mongo lol
type Store struct {
	FBIDMap      map[string]*Entry
	NameMap      map[string]*Entry
	ChannelIDMap map[string]*Entry
}

type Entry struct {
	FBID      string
	Name      string
	ChannelID string
}

func (T *Store) getByFBID(fbid string) (*Entry, error) {
	if entry, ok := T.FBIDMap[fbid]; ok {
		return entry, nil
	}
	return nil, errors.New("entry not found")
}

func (T *Store) getByName(name string) (*Entry, error) {
	if entry, ok := T.NameMap[name]; ok {
		return entry, nil
	}
	return nil, errors.New("entry not found")
}

func (T *Store) getByChannelID(channelID string) (*Entry, error) {
	if entry, ok := T.ChannelIDMap[channelID]; ok {
		return entry, nil
	}
	return nil, errors.New("entry not found")
}

func (T *Store) upsertEntry(entry *Entry) {
	T.FBIDMap[entry.FBID] = entry
	T.NameMap[entry.Name] = entry
	T.ChannelIDMap[entry.ChannelID] = entry
}

func (T *Store) deleteEntry(entry *Entry) {
	// TODO: implement
}
