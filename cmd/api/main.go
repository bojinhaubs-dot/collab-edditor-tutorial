package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"codex-hubcoding-test/internal/httpapi"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Go API listening on http://%s", addr)
	log.Fatal(server.ListenAndServe())
}
