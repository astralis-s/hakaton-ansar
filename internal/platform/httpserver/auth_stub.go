package httpserver

import (
	"log/slog"
	"net/http"
)

// NOTE (Phase 1): the two functions below are deliberate placeholders so the two
// HTTP surfaces and their middleware wiring already exist. They currently pass
// every request through. They will be replaced by real implementations:
//   - StubBearerAuth  → JWT validation from the iam module          (Phase 2)
//   - StubAPIKeyAuth   → X-API-Key validation (iam/publicapi auth)  (Phase 2/6)

// StubBearerAuth is the Phase 1 placeholder for the internal /api/app surface
// (eventual JWT "Authorization: Bearer <token>" check).
func StubBearerAuth(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO(phase2): validate JWT via iam and inject the authenticated user.
			next.ServeHTTP(w, r)
		})
	}
}

// StubAPIKeyAuth is the Phase 1 placeholder for the public /api/v1 surface
// (eventual "X-API-Key" check).
func StubAPIKeyAuth(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO(phase2): validate X-API-Key and inject the owning organization.
			next.ServeHTTP(w, r)
		})
	}
}
