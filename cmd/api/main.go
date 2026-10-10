package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"codex-hubcoding-test/internal/database"
	"codex-hubcoding-test/internal/httpapi"
	"codex-hubcoding-test/internal/model"
	"codex-hubcoding-test/internal/repository"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN is required")
	}
	db, err := database.Open(dsn)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get sql db: %v", err)
	}
	defer sqlDB.Close()
	log.Println("MySQL connected successfully")
	if err := db.AutoMigrate(&model.Document{}); err != nil {
		log.Fatalf("migrate documents table failed: %v", err)
	}
	log.Println("Document table migration completed")

	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewRouter(repository.NewDocuments(db)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Go API listening on http://%s", addr)
	log.Fatal(server.ListenAndServe())
}
