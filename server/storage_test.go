package main

import (
	"testing"
	"time"
)

var store *inMem

func TestInMemStorageGet(t *testing.T) {
	store = NewInMemStorage()

	// test set
	cid := "dummyclient"
	state := &SessionData{
		ClientID: cid,
	}
	store.Store(state)
	st, err := store.Retrive(cid)
	if err != nil || st.ClientID != cid {
		t.Error("key not found")
	}

	state.DiscardTs = time.Now().Unix() - 40
	store.Store(state)
	st, err = store.Retrive(cid)
	if err == nil {
		t.Error("key was not expired")
	}
}
