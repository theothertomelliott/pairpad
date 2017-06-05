package main

import (
	"strconv"
	"sync"
	"time"
)

type Messaging struct {
	StartTime      time.Time
	Incoming       chan message
	sessionsByName map[string]Session
	sessionMtx     sync.Mutex
	SessionRequest chan chan Session
	nextSession    int
}

func NewMessaging() *Messaging {
	return &Messaging{
		Incoming:       make(chan message),
		SessionRequest: make(chan chan Session),
		sessionsByName: make(map[string]Session),
	}
}

func (m *Messaging) GetSession(sessionId string) Session {
	m.sessionMtx.Lock()
	defer m.sessionMtx.Unlock()
	return m.sessionsByName[sessionId]
}

func (m *Messaging) Run() {
	for true {
		select {
		case msg := <-m.Incoming:
			for index, delta := range msg.Deltas {
				delta.TimeOffset = time.Since(m.StartTime)
				msg.Deltas[index] = delta
			}
			m.sessionMtx.Lock()
			for _, session := range m.sessionsByName {
				if msg.SessionId != session.Id {
					// TODO: Mark this session as failed if needed
					session.Messages <- msg
				}
			}
			m.sessionMtx.Unlock()
		case requestChan := <-m.SessionRequest:
			newSession := Session{
				Id:       strconv.Itoa(m.nextSession),
				Messages: make(chan message, 100), // TODO: Determine an appropriate buffer size
			}
			m.nextSession++
			requestChan <- newSession
			m.sessionsByName[newSession.Id] = newSession
		}
	}
}

type Session struct {
	Id       string
	Messages chan message
}

type pos struct {
	Row    int `json:"row"`
	Column int `json:"column"`
}

type delta struct {
	Start      pos           `json:"start"`
	End        pos           `json:"end"`
	Lines      []string      `json:"lines"`
	Action     string        `json:"action"`
	TimeOffset time.Duration `json:"time"`
}

type message struct {
	SessionId string
	Deltas    []delta `json:"deltas"`
}
