package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var echoUpgrader = websocket.Upgrader{}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /ws/echo", echo)
	return mux
}

func echo(w http.ResponseWriter, r *http.Request) {
	conn, err := echoUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade writes the HTTP error response.
	}
	defer conn.Close()
	conn.SetReadLimit(1 << 20)

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, payload); err != nil {
			return
		}
	}
}
