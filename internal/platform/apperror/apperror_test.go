package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"plain error → 500", errors.New("boom"), http.StatusInternalServerError},
		{"internal → 500", Internal("x", errors.New("y")), http.StatusInternalServerError},
		{"invalid → 400", Invalid("bad_input", "bad"), http.StatusBadRequest},
		{"unauthorized → 401", Unauthorized("no_token", "no token"), http.StatusUnauthorized},
		{"forbidden → 403", Forbidden("not_owner", "owner only"), http.StatusForbidden},
		{"not found → 404", NotFound("contract", "not found"), http.StatusNotFound},
		{"conflict → 409", Conflict("dup", "exists"), http.StatusConflict},
		{"wrapped invalid → 400", fmt.Errorf("ctx: %w", Invalid("bad", "bad")), http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := HTTPStatus(tc.err); got != tc.want {
				t.Fatalf("HTTPStatus(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

func TestWrite_InternalDoesNotLeakDetail(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	Write(rec, req, nil, Internal("db exploded with secret dsn", errors.New("pq: password=hunter2")))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if want := `"internal error"`; !contains(body, want) {
		t.Fatalf("body %q must contain generic message %q", body, want)
	}
	if contains(body, "hunter2") || contains(body, "db exploded") {
		t.Fatalf("body leaked internal detail: %q", body)
	}
}

func TestWrite_ClassifiedMessageIsExposed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	Write(rec, req, nil, NotFound("contract_not_found", "договор не найден"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if body := rec.Body.String(); !contains(body, "contract_not_found") {
		t.Fatalf("body %q must contain code", body)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
