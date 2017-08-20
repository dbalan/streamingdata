package main

import (
	"fmt"
	"sync"
	"time"
)

type SessionData struct {
	ClientID  string
	Count     int64
	DiscardTs int64
	Seed      int64
}

type inMem struct {
	// both accessors modify map.
	lock    sync.Mutex
	storage map[string]*SessionData
}

type Storage interface {
	Store(*SessionData) error
	Retrive(string) (*SessionData, error)
}

func NewInMemStorage() *inMem {
	s := make(map[string]*SessionData)
	return &inMem{sync.Mutex{}, s}
}

func (s *inMem) Store(sd *SessionData) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.storage[sd.ClientID] = sd
	return nil
}

func (s *inMem) Retrive(cid string) (*SessionData, error) {
	s.lock.Lock()
	state, ok := s.storage[cid]
	delete(s.storage, cid)
	s.lock.Unlock()
	if !ok {
		return nil, fmt.Errorf("no such key")
	}
	if state.DiscardTs != 0 {
		if time.Now().Unix()-30 >= state.DiscardTs {
			return nil, fmt.Errorf("key expired")
		}
	}
	return state, nil
}
