package main

import (
	"time"
)

type Document struct {
	*Sessions
	Chat *Chat

	StartTime time.Time
	Incoming  chan update

	pendingRequests []UpdateRequest
	updateHistory   []update

	UpdateRequest chan UpdateRequest
}

func NewDocument() *Document {
	return &Document{
		Sessions:      NewSessions(),
		Chat:          NewChat(),
		Incoming:      make(chan update),
		UpdateRequest: make(chan UpdateRequest),
	}
}

func (m *Document) Run() {
	go m.Sessions.Run()
	go m.Chat.Run()

	receiverTimeout := time.Tick(10 * time.Second)
	for true {
		select {
		case msg := <-m.Incoming:
			m.handleIncoming(msg)
		case request := <-m.UpdateRequest:
			m.handleMessageRequest(request)
		case <-receiverTimeout: // Close all receivers to prevent long poll timeouts
			m.handleTimeout()
		}
	}
}

func (m *Document) handleIncoming(msg update) {
	msg.TimeOffset = time.Since(m.StartTime)
	msg.UpdateID = len(m.updateHistory)
	for _, request := range m.pendingRequests {
		request.Receiver <- &msg
		close(request.Receiver)
	}
	m.pendingRequests = make([]UpdateRequest, 0)

	m.updateHistory = append(m.updateHistory, msg)
}

func (m *Document) handleMessageRequest(request UpdateRequest) {
	if request.FirstMessage < len(m.updateHistory) {
		for _, msg := range m.updateHistory[request.FirstMessage:] {
			msg := msg
			request.Receiver <- &msg
		}
		close(request.Receiver)
	} else {
		m.pendingRequests = append(m.pendingRequests, request)
	}
}

func (m *Document) handleTimeout() {
	for _, request := range m.pendingRequests {
		close(request.Receiver)
	}
	m.pendingRequests = make([]UpdateRequest, 0)
}

type UpdateRequest struct {
	FirstMessage int
	Receiver     chan *update
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

type update struct {
	SessionID          string        `json:"sessionId"`
	UpdateID           int           `json:"messageId"`
	TimeOffset         time.Duration `json:"time"`
	Delta              delta         `json:"delta"`
	LanguageSelections string        `json:"languageSelection"`
}
