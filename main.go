package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var documents *DocumentPool

var fileServer http.Handler

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	fileServer = http.FileServer(http.Dir("public/"))

	documents = NewDocumentPool()
	go documents.Run()

	http.HandleFunc("/chat/poll/", ChatPollResponse)
	http.HandleFunc("/chat/push/", ChatPushHandler)

	http.HandleFunc("/poll/", PollResponse)
	http.HandleFunc("/push/", PushHandler)
	http.HandleFunc("/doc/new", NewDocumentHandler)
	http.HandleFunc("/doc/", DocumentHandler)
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
		fmt.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func NewDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getNewDocumentID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/doc/%s", id), http.StatusFound)
}

func DocumentHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	documentID := strings.Replace(r.URL.Path, "/doc/", "", 1)
	messaging, err := getDocument(documentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t := template.New("document.html")
	t, err = t.ParseFiles("views/document.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := make(chan Session)
	messaging.SessionRequest <- c

	session := <-c

	fmt.Printf("Started session %s in document %s", session.Id, documentID)

	messaging.Chat.Incoming <- chatUpdate{
		SessionNameChanges: map[string]string{
			session.Id: fmt.Sprintf("Guest %s", session.Id),
		},
	}

	err = t.Execute(w, map[string]interface{}{
		"DocumentId": documentID,
		"SessionId":  session.Id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func PollResponse(w http.ResponseWriter, req *http.Request) {
	var err error
	documentID := strings.Replace(req.URL.Path, "/poll/", "", 1)
	messaging, err := getDocument(documentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	receiver := make(chan *update)
	messaging.UpdateRequest <- UpdateRequest{
		FirstMessage: nextMessageInt,
		SessionID:    sessionID,
		Receiver:     receiver,
	}

	var messages []*update
	for msg := range receiver {
		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		io.WriteString(w, "[]")
		return
	}

	content, err := json.Marshal(messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	io.WriteString(w, string(content))
}

func PushHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	documentID := strings.Replace(req.URL.Path, "/push/", "", 1)
	messaging, err := getDocument(documentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var e = update{}
	err = decoder.Decode(&e)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	messaging.Incoming <- e
}

func ChatPollResponse(w http.ResponseWriter, req *http.Request) {
	var err error
	documentID := strings.Replace(req.URL.Path, "/chat/poll/", "", 1)
	messaging, err := getDocument(documentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	receiver := make(chan *chatUpdate)
	messaging.Chat.UpdateRequest <- ChatUpdateRequest{
		FirstMessage: nextMessageInt,
		SessionID:    sessionID,
		Receiver:     receiver,
	}

	var messages []*chatUpdate
	for msg := range receiver {
		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		io.WriteString(w, "[]")
		return
	}

	content, err := json.Marshal(messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	io.WriteString(w, string(content))
}

func ChatPushHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	documentID := strings.Replace(req.URL.Path, "/chat/push/", "", 1)
	messaging, err := getDocument(documentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var e = chatUpdate{}
	err = decoder.Decode(&e)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	messaging.Chat.Incoming <- e
}

func getNewDocumentID() (string, error) {
	responseChan := make(chan NewDocumentResponse)
	documents.New <- responseChan
	select {
	case response := <-responseChan:
		return response.Id, nil
	case <-time.After(1 * time.Second):
		return "", errors.New("timed out waiting for new session")
	}
}

func getDocument(id string) (*Document, error) {
	request := DocumentRequest{
		Id:       id,
		Response: make(chan *Document),
		Error:    make(chan error),
	}
	documents.Existing <- request
	select {
	case response := <-request.Response:
		return response, nil
	case err := <-request.Error:
		return nil, err
	case <-time.After(1 * time.Second):
		return nil, errors.New("timed out waiting for existing session")
	}
}
