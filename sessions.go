package main

import "strconv"

type Sessions struct {
	SessionRequest chan chan Session
	nextSession    int
}

func NewSessions() *Sessions {
	return &Sessions{
		SessionRequest: make(chan chan Session),
	}
}

func (s *Sessions) Run() {
	for requestChan := range s.SessionRequest {
		s.handleSessionRequest(requestChan)
	}
}

func (s *Sessions) handleSessionRequest(requestChan chan Session) {
	newSession := Session{
		Id: strconv.Itoa(s.nextSession),
	}
	s.nextSession++
	requestChan <- newSession
}

type Session struct {
	Id string
}
