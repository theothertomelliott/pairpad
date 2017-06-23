package main

import "time"

type Chat struct {
	sessionNames map[string]string

	pendingRequests []ChatUpdateRequest
	updateHistory   []chatUpdate

	Incoming       chan chatUpdate
	SetNameRequest chan SetNameRequest
	UpdateRequest  chan ChatUpdateRequest
}

func NewChat() *Chat {
	return &Chat{
		Incoming:       make(chan chatUpdate),
		SetNameRequest: make(chan SetNameRequest),
		UpdateRequest:  make(chan ChatUpdateRequest),
		sessionNames:   make(map[string]string),
	}
}

func (c *Chat) Run() {
	receiverTimeout := time.Tick(10 * time.Second)
	for true {
		select {
		case msg := <-c.Incoming:
			c.handleIncoming(msg)
		case request := <-c.UpdateRequest:
			c.handleMessageRequest(request)
		case <-receiverTimeout: // Close all receivers to prevent long poll timeouts
			c.handleTimeout()
		}
	}
}

func (c *Chat) handleIncoming(msg chatUpdate) {
	msg.UpdateID = len(c.updateHistory)
	for _, request := range c.pendingRequests {
		request.Receiver <- &msg
		close(request.Receiver)
	}
	c.pendingRequests = make([]ChatUpdateRequest, 0)
	c.updateHistory = append(c.updateHistory, msg)
}

func (c *Chat) handleMessageRequest(request ChatUpdateRequest) {
	if request.FirstMessage < len(c.updateHistory) {
		for _, msg := range c.updateHistory[request.FirstMessage:] {
			msg := msg
			request.Receiver <- &msg
		}
		close(request.Receiver)
	} else {
		c.pendingRequests = append(c.pendingRequests, request)
	}
}

func (c *Chat) handleTimeout() {
	for _, request := range c.pendingRequests {
		close(request.Receiver)
	}
	c.pendingRequests = make([]ChatUpdateRequest, 0)
}

type ChatMessage struct {
	SessionID string
	Message   string
}

type ChatUpdateRequest struct {
	FirstMessage int
	SessionID    string
	Receiver     chan *chatUpdate
}

type SetNameRequest struct {
	SessionID string
	Name      string
	Response  chan struct{}
}

type chatUpdate struct {
	UpdateID           int               `json:"messageId"`
	Time               time.Time         `json:"time"`
	Messages           []ChatMessage     `json:"messages"`
	SessionNameChanges map[string]string `json:"sessionNameChanges"`
	SessionsQuit       []string          `json:"sessionsQuit"`
}
