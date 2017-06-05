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

func TestMessaging(t *testing.T) {
	sessionCount := 3
	messaging := NewMessaging()
	go messaging.Run()
	var sessions []Session
	for i := 0; i < sessionCount; i++ {
		c := make(chan Session)
		messaging.SessionRequest <- c
		select {
		case session := <-c:
			sessions = append(sessions, session)
		case <-time.After(1 * time.Second):
			t.Errorf("Timed out waiting for session")
		}
	}

	for _, session := range sessions {
		messaging.Incoming <- message{
			SessionId: session.Id,
		}
		for _, receiver := range sessions {
			if receiver.Id == session.Id {
				continue
			}
			select {
			case m := <-receiver.Messages:
				if m.SessionId != session.Id {
					t.Error("SessionId not as expected")
				}
			case <-time.After(1 * time.Second):
				t.Errorf("Timed out waiting for message from session %v, to session: %v", session.Id, receiver.Id)
			}
		}

		select {
		case _ = <-session.Messages:
			t.Error("Session should not have received message")
		default:
		}
	}

}
