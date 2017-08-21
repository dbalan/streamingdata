package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v5"
	"sync"
	"time"
)

var KeyExpiredError = fmt.Errorf("key expired")
var AlreadyPresentError = fmt.Errorf("session duplicate")

type SessionData struct {
	ClientID  string `json:"client_id"`
	Last      int64  `json:"last"`
	DiscardTs int64  `json:"discard_ts"`
	Seed      int64  `json:"seed"`
}

type inMem struct {
	// both accessors modify map.
	lock    sync.Mutex
	storage map[string]*SessionData
	lockmap map[string]bool
}

type redisMem struct {
	client *redis.Client
}

type Storage interface {
	Store(*SessionData) error
	Retrive(string) (*SessionData, error)
}

func NewRedisStorage() *redisMem {
	// assumes redis running on localhost
	// FIXME: move to cmd
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	client.Ping()
	return &redisMem{client}
}

func (r *redisMem) Store(sd *SessionData) error {
	blob, err := json.Marshal(sd)
	if err != nil {
		return err
	}
	return r.client.Set(sd.ClientID, blob, 0).Err()
}

func (r *redisMem) Retrive(cid string) (*SessionData, error) {
	pipe := r.client.TxPipeline()

	blob := pipe.Get(cid)
	pipe.Del(cid)
	_, err := pipe.Exec()
	if err != nil || blob.Err() != nil {
		return nil, err
	}

	var s SessionData
	err = json.Unmarshal([]byte(blob.Val()), &s)
	if err != nil {
		return nil, err
	}
	if s.DiscardTs != 0 {
		if time.Now().Unix()-30 >= s.DiscardTs {
			return nil, KeyExpiredError
		}
	}
	return &s, nil
}

func NewInMemStorage() *inMem {
	s := make(map[string]*SessionData)
	l := make(map[string]bool)
	return &inMem{sync.Mutex{}, s, l}
}

func (s *inMem) Store(sd *SessionData) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.storage[sd.ClientID] = sd
	delete(s.lockmap, sd.ClientID)
	return nil
}

func (s *inMem) Retrive(cid string) (*SessionData, error) {
	s.lock.Lock()
	_, ok := s.lockmap[cid]
	if ok {
		return nil, AlreadyPresentError
	}
	state, ok := s.storage[cid]
	delete(s.storage, cid)
	s.lockmap[cid] = true
	s.lock.Unlock()
	if !ok {
		return nil, fmt.Errorf("no such key")
	}
	if state.DiscardTs != 0 {
		if time.Now().Unix()-30 >= state.DiscardTs {
			return nil, KeyExpiredError
		}
	}
	return state, nil
}
