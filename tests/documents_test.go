package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"codex-hubcoding-test/internal/httpapi"
)

type documentResponse struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func TestCreateAndGetDocument(t *testing.T) {
	router := httpapi.NewRouter(newTestRepository())
	created := httptest.NewRecorder()
	router.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(`{"title":"  main.go  ","content":"package main\n"}`)))
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", created.Code, http.StatusCreated, created.Body.String())
	}

	var doc documentResponse
	if err := json.Unmarshal(created.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ID == "" || doc.Title != "main.go" || doc.Content != "package main\n" {
		t.Fatalf("created document = %+v", doc)
	}

	found := httptest.NewRecorder()
	router.ServeHTTP(found, httptest.NewRequest(http.MethodGet, "/documents/"+doc.ID, nil))
	if found.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", found.Code, http.StatusOK)
	}
	var fetched documentResponse
	if err := json.Unmarshal(found.Body.Bytes(), &fetched); err != nil {
		t.Fatal(err)
	}
	if fetched != doc {
		t.Fatalf("fetched document = %+v, want %+v", fetched, doc)
	}
}

func TestGetMissingDocument(t *testing.T) {
	response := httptest.NewRecorder()
	httpapi.NewRouter(newTestRepository()).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/documents/999", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestGetInvalidDocumentID(t *testing.T) {
	response := httptest.NewRecorder()
	httpapi.NewRouter(newTestRepository()).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/documents/not-a-number", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestCreateDocumentBadRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"malformed JSON", `{"title":`},
		{"missing title", `{"content":""}`},
		{"blank title", `{"title":"   ","content":""}`},
		{"missing content", `{"title":"main.go"}`},
		{"wrong content type", `{"title":"main.go","content":123}`},
	}

	router := httpapi.NewRouter(newTestRepository())
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(test.body)))
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
		})
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(`{"title":"main.go","content":""}`)))
	if response.Code != http.StatusCreated {
		t.Fatalf("valid request after bad requests returned %d, want %d", response.Code, http.StatusCreated)
	}
}

func TestConcurrentDocumentRequests(t *testing.T) {
	const count = 64
	router := httpapi.NewRouter(newTestRepository())
	ids := make(chan string, count)
	errors := make(chan error, count)
	var workers sync.WaitGroup

	for range count {
		workers.Add(1)
		go func() {
			defer workers.Done()
			created := httptest.NewRecorder()
			router.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(`{"title":"main.go","content":""}`)))
			if created.Code != http.StatusCreated {
				errors <- fmt.Errorf("create status = %d", created.Code)
				return
			}

			var doc documentResponse
			if err := json.Unmarshal(created.Body.Bytes(), &doc); err != nil {
				errors <- err
				return
			}
			fetched := httptest.NewRecorder()
			router.ServeHTTP(fetched, httptest.NewRequest(http.MethodGet, "/documents/"+doc.ID, nil))
			if fetched.Code != http.StatusOK {
				errors <- fmt.Errorf("get %q status = %d", doc.ID, fetched.Code)
				return
			}
			ids <- doc.ID
		}()
	}
	workers.Wait()
	close(ids)
	close(errors)

	for err := range errors {
		t.Error(err)
	}
	seen := make(map[string]bool)
	for id := range ids {
		if id == "" || seen[id] {
			t.Errorf("missing or duplicate document ID %q", id)
		}
		seen[id] = true
	}
	if len(seen) != count {
		t.Errorf("created %d distinct documents, want %d", len(seen), count)
	}
}
