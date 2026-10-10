package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"codex-hubcoding-test/internal/database"
	"codex-hubcoding-test/internal/httpapi"
	"codex-hubcoding-test/internal/model"
	"codex-hubcoding-test/internal/repository"
)

type failureRepository struct{}

func (failureRepository) Create(context.Context, *model.Document) error {
	return errors.New("database unavailable")
}

func (failureRepository) GetByID(context.Context, uint64) (model.Document, error) {
	return model.Document{}, errors.New("database unavailable")
}

func TestDatabaseFailureResponses(t *testing.T) {
	router := httpapi.NewRouter(failureRepository{})
	for _, test := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/documents", `{"title":"main.go","content":""}`},
		{http.MethodGet, "/documents/1", ""},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(test.method, test.path, strings.NewReader(test.body)))
		if response.Code != http.StatusInternalServerError {
			t.Errorf("%s %s status = %d, want 500", test.method, test.path, response.Code)
		}
	}
}

func TestDatabaseConnectionFailure(t *testing.T) {
	for _, test := range []struct {
		name string
		dsn  string
	}{
		{"invalid DSN", "invalid DSN"},
		{"unreachable server", "test:test@tcp(127.0.0.1:1)/collab?timeout=1s"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := database.Open(test.dsn); err == nil {
				t.Fatal("expected a database connection error")
			}
		})
	}
}

func TestMySQLDocumentSurvivesReconnect(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("set MYSQL_TEST_DSN to run the MySQL integration test")
	}

	db, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Document{}); err != nil {
		t.Fatal(err)
	}
	router := httpapi.NewRouter(repository.NewDocuments(db))
	created := httptest.NewRecorder()
	router.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(`{"title":"restart.go","content":"package main"}`)))
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d; body = %s", created.Code, created.Body.String())
	}
	var doc documentResponse
	if err := json.Unmarshal(created.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}

	// A new connection and router must find the row without relying on process memory.
	reopened, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	reopenedSQL, err := reopened.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		reopened.Delete(&model.Document{}, doc.ID)
		reopenedSQL.Close()
	}()
	newRouter := httpapi.NewRouter(repository.NewDocuments(reopened))
	found := httptest.NewRecorder()
	newRouter.ServeHTTP(found, httptest.NewRequest(http.MethodGet, "/documents/"+doc.ID, nil))
	if found.Code != http.StatusOK {
		t.Fatalf("get after reconnect status = %d; body = %s", found.Code, found.Body.String())
	}
	var fetched documentResponse
	if err := json.Unmarshal(found.Body.Bytes(), &fetched); err != nil {
		t.Fatal(err)
	}
	if fetched != doc {
		t.Fatalf("fetched document = %+v, want %+v", fetched, doc)
	}
}
