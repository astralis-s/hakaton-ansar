package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam"
)

// newTestRouter builds the real route tree. The iam module is constructed with a
// nil pool: the routes exercised here (health, swagger, and the auth-rejection
// paths) never reach the database.
func newTestRouter() chi.Router {
	iamModule := iam.New(iam.Deps{
		Pool:      nil,
		Tx:        nil,
		Log:       nil,
		JWTSecret: "test-secret",
		JWTTTL:    0,
	})
	r := chi.NewRouter()
	mountRoutes(r, iamModule)
	return r
}

func TestPublicRoutes(t *testing.T) {
	r := newTestRouter()

	cases := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"health", http.MethodGet, "/health", http.StatusOK},
		{"swagger doc", http.MethodGet, "/swagger/doc.json", http.StatusOK},
		{"protected app route without token → 401", http.MethodGet, "/api/app/auth/me", http.StatusUnauthorized},
		{"public api route without key → 401", http.MethodGet, "/api/v1/ping", http.StatusUnauthorized},
		{"unknown route → 404", http.MethodGet, "/nope", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("%s %s = %d, want %d (body: %s)", tc.method, tc.path, rec.Code, tc.wantStatus, rec.Body.String())
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
