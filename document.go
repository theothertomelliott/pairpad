package main

import (
	"strconv"
	"time"
)

type Document struct {
	StartTime time.Time
	Incoming  chan message

	pendingRequests []MessageRequest
	messageHistory  []message

	SessionRequest chan chan Session
	MessageRequest chan MessageRequest
	nextSession    int
}

func NewDocument() *Document {
	return &Document{
		Incoming:       make(chan message),
		SessionRequest: make(chan chan Session),
		MessageRequest: make(chan MessageRequest),
	}
}

func (m *Document) Run() {
	receiverTimeout := time.Tick(10 * time.Second)
	for true {
		select {
		case msg := <-m.Incoming:
			msg.TimeOffset = time.Since(m.StartTime)
			msg.MessageID = len(m.messageHistory)
			for _, request := range m.pendingRequests {
				request.Receiver <- &msg
				close(request.Receiver)
			}
			m.pendingRequests = make([]MessageRequest, 0)

			m.messageHistory = append(m.messageHistory, msg)

		case requestChan := <-m.SessionRequest:
			newSession := Session{
				Id:       strconv.Itoa(m.nextSession),
				Messages: make(chan message, 100), // TODO: Determine an appropriate buffer size
			}
			m.nextSession++
			requestChan <- newSession
		case request := <-m.MessageRequest:
			if request.FirstMessage < len(m.messageHistory) {
				for _, msg := range m.messageHistory[request.FirstMessage:] {
					msg := msg
					request.Receiver <- &msg
				}
				close(request.Receiver)
			} else {
				m.pendingRequests = append(m.pendingRequests, request)
			}
		case <-receiverTimeout: // Close all receivers to prevent long poll timeouts
			for _, request := range m.pendingRequests {
				close(request.Receiver)
			}
			m.pendingRequests = make([]MessageRequest, 0)
		}

	}
}

type MessageRequest struct {
	FirstMessage int
	SessionID    string
	Receiver     chan *message
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
	Start  pos      `json:"start"`
	End    pos      `json:"end"`
	Lines  []string `json:"lines"`
	Action string   `json:"action"`
}

type message struct {
	SessionID  string        `json:"sessionId"`
	MessageID  int           `json:"messageId"`
	TimeOffset time.Duration `json:"time"`
	Deltas     []delta       `json:"deltas"`
}
