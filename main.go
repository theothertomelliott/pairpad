package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("public/")))
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Loading template at URL:", r.URL.Path)
		var err error
		t := template.New("index.html")
		t, err = t.ParseFiles("views/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})
	http.HandleFunc("/poll", PollResponse)
	http.HandleFunc("/push", PushHandler)
	http.HandleFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		close(messages)
	})

	fmt.Println("Starting to listen on port 51936")
	http.ListenAndServe(":51936", nil)
}

var messages chan string = make(chan string)

func PollResponse(w http.ResponseWriter, req *http.Request) {
	for msg := range messages {
		io.WriteString(w, msg)
	}
}

func PushHandler(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if msg, ok := req.Form["msg"]; ok {
		messages <- string(msg[0])
		return
	}
	http.Error(w, "msg is required", http.StatusBadRequest)
}
