package main

import (
	"strconv"
	"testing"
	"time"
)

func TestGetSession(t *testing.T) {
	messaging := NewMessaging()
	go messaging.Run()
	for i := 0; i < 10; i++ {
		c := make(chan Session)
		messaging.SessionRequest <- c
		select {
		case session := <-c:
			if session.Id != strconv.Itoa(i) {
				t.Errorf("Wrong id received")
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Timed out waiting for session")
		}
	}
}

func TestMessageRequestCatchUp(t *testing.T) {
	messaging := NewMessaging()
	go messaging.Run()

	for i := 0; i < 2; i++ {
		messaging.Incoming <- message{
			SessionID: "session",
		}
	}

	receiver := make(chan *message)
	messaging.MessageRequest <- MessageRequest{
		FirstMessage: 0,
		SessionID:    "another_session",
		Receiver:     receiver,
	}

	var count int
	for m := range receiver {
		if m.SessionID != "session" {
			t.Error("SessionId not as expected:", m.SessionID)
		}
		if m.MessageID != count {
			t.Error("MessageId not as expected:", m.MessageID)
		}
		count++
	}
	if count != 2 {
		t.Error("Expected two messages")
	}
}

func TestMessageRequestPending(t *testing.T) {
	messaging := NewMessaging()
	go messaging.Run()

	receiver := make(chan *message)
	messaging.MessageRequest <- MessageRequest{
		FirstMessage: 0,
		SessionID:    "another_session",
		Receiver:     receiver,
	}

	messaging.Incoming <- message{
		SessionID: "session",
	}

	var count int
	for m := range receiver {
		if m.SessionID != "session" {
			t.Error("SessionId not as expected:", m.SessionID)
		}
		if m.MessageID != count {
			t.Error("MessageId not as expected:", m.MessageID)
		}
		count++
	}
	if count != 1 {
		t.Error("Expected one message")
	}
}
