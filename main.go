package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

var messaging *Messaging
var sessions map[string]Session

func main() {
	messaging = NewMessaging()
	go messaging.Run()

	sessions = make(map[string]Session)

	http.HandleFunc("/index.html", Index)
	http.HandleFunc("/poll", PollResponse)
	http.HandleFunc("/push", PushHandler)
	http.Handle("/", http.FileServer(http.Dir("public/")))

	fmt.Println("Starting to listen on port 51936")
	http.ListenAndServe(":51936", nil)
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Loading template at URL:", r.URL.Path)
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
	// TODO: Validate
	sessionId := req.Form["sessionId"][0]
	session := messaging.GetSession(sessionId)
	msg := <-session.Messages
	content, _ := json.Marshal(msg)
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
