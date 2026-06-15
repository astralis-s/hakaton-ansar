package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestRouter() chi.Router {
	r := chi.NewRouter()
	mountRoutes(r, testLogger())
	return r
}

func TestRoutesReturn200(t *testing.T) {
	r := newTestRouter()

	cases := []struct {
		path       string
		wantStatus int
	}{
		{"/health", http.StatusOK},
		{"/api/app/ping", http.StatusOK},
		{"/api/v1/ping", http.StatusOK},
		{"/swagger/doc.json", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("GET %s = %d, want %d (body: %s)", tc.path, rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHealthBody(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %q, want \"ok\"", body["status"])
	}
}

func TestSwaggerDocIsJSON(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var spec map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("swagger spec is not valid JSON: %v", err)
	}
	if spec["swagger"] != "2.0" {
		t.Fatalf("swagger version = %v, want 2.0", spec["swagger"])
	}
}
