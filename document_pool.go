package main

import (
	"errors"
	"fmt"
	"strconv"
)

type DocumentPool struct {
	messaging map[string]*Document
	New       chan chan NewDocumentResponse
	Existing  chan DocumentRequest
}

func NewDocumentPool() *DocumentPool {
	return &DocumentPool{
		messaging: make(map[string]*Document),
		New:       make(chan chan NewDocumentResponse),
		Existing:  make(chan DocumentRequest),
	}
}

func (m *DocumentPool) Run() {
	for true {
		select {
		case msg := <-m.New:
			fmt.Println("Request for new")
			id := strconv.Itoa(len(m.messaging) + 1)
			messaging := NewDocument()
			go messaging.Run()
			m.messaging[id] = messaging
			msg <- NewDocumentResponse{
				Id:       id,
				Document: messaging,
			}
			close(msg)
		case msg := <-m.Existing:
			messaging, ok := m.messaging[msg.Id]
			if ok {
				msg.Response <- messaging
			} else {
				msg.Error <- errors.New("document not found")
			}
			close(msg.Response)
			close(msg.Error)
		}
	}
}

type NewDocumentResponse struct {
	Id       string
	Document *Document
}

type DocumentRequest struct {
	Id       string
	Response chan *Document
	Error    chan error
}
