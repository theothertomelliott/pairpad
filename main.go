package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var messaging *Messaging
var sessions map[string]Session

var fileServer http.Handler

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	fileServer = http.FileServer(http.Dir("public/"))

	messaging = NewMessaging()
	go messaging.Run()

	sessions = make(map[string]Session)

	http.HandleFunc("/poll", PollResponse)
	http.HandleFunc("/push", PushHandler)
	http.HandleFunc("/", Index)

	fmt.Println("Starting to listen on port", port)
	http.ListenAndServe(":"+port, nil)
}

func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		fileServer.ServeHTTP(w, r)
		return
	}

	var err error
	t := template.New("index.html")
	t, err = t.ParseFiles("views/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := make(chan Session)
	messaging.SessionRequest <- c

	session := <-c
	sessions[session.Id] = session

	fmt.Println("Started session: ", session.Id)

	err = t.Execute(w, map[string]interface{}{
		"SessionId": session.Id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func PollResponse(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	if s, ok := req.Form["sessionId"]; !ok || len(s) == 0 {
		http.Error(w, "sessionId is required", http.StatusBadRequest)
		return
	}
	if n, ok := req.Form["next"]; !ok || len(n) == 0 {
		http.Error(w, "next message number is required", http.StatusBadRequest)
		return
	}

	sessionID := req.Form["sessionId"][0]
	nextMessageStr := req.Form["next"][0]
	nextMessageInt, err := strconv.Atoi(nextMessageStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	receiver := make(chan *message)
	messaging.MessageRequest <- MessageRequest{
		FirstMessage: nextMessageInt,
		SessionID:    sessionID,
		Receiver:     receiver,
	}

	var messages []*message
	for msg := range receiver {
		messages = append(messages, msg)
	}

	content, _ := json.Marshal(messages)
	io.WriteString(w, string(content))
}

func PushHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var e = message{}
	err := decoder.Decode(&e)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	messaging.Incoming <- e
}
