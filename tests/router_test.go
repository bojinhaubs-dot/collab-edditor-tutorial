package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"codex-hubcoding-test/internal/httpapi"
)

func TestHealth(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	httpapi.NewRouter(newTestRepository()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", response.Body.String(), `{"status":"ok"}`)
	}
	if response.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
	}
}
