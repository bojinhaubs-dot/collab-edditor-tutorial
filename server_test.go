package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestHealthz(t *testing.T) {
	server := httptest.NewServer(newHandler())
	defer server.Close()

	response, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(body) != `{"status":"ok"}` {
		t.Fatalf("healthz: status = %d, body = %q", response.StatusCode, body)
	}
}

func TestWebSocketEcho(t *testing.T) {
	server := httptest.NewServer(newHandler())
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/echo"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for _, test := range []struct {
		name        string
		messageType int
		payload     string
	}{
		{"JSON text", websocket.TextMessage, `{"type":"ping","text":"hello"}`},
		{"binary", websocket.BinaryMessage, "\x00\x01\x02"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := conn.WriteMessage(test.messageType, []byte(test.payload)); err != nil {
				t.Fatal(err)
			}
			messageType, payload, err := conn.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}
			if messageType != test.messageType || string(payload) != test.payload {
				t.Fatalf("echo: type = %d, payload = %q", messageType, payload)
			}
		})
	}
}
