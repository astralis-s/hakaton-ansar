package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
	"github.com/astralis-s/hakaton-ansar/internal/platform/authctx"
)

// JWTAuth validates "Authorization: Bearer <jwt>" for the internal /api/app
// surface and attaches the principal to the request context.
func JWTAuth(tokens domain.TokenService, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := bearerToken(r)
			if raw == "" {
				apperror.Write(w, r, log, apperror.Unauthorized("missing_token", "missing bearer token"))
				return
			}
			principal, err := tokens.Parse(raw)
			if err != nil {
				apperror.Write(w, r, log, apperror.Unauthorized("invalid_token", "invalid or expired token"))
				return
			}
			ctx := authctx.With(r.Context(), authctx.Principal{
				UserID: principal.UserID,
				OrgID:  principal.OrgID,
				Role:   principal.Role.String(),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// APIKeyAuth validates the "X-API-Key" header for the public /api/v1 surface and
// attaches the owning organization (no user) to the context.
func APIKeyAuth(authenticator *app.AuthenticateApiKey, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := strings.TrimSpace(r.Header.Get("X-API-Key"))
			if key == "" {
				apperror.Write(w, r, log, apperror.Unauthorized("missing_api_key", "missing X-API-Key header"))
				return
			}
			orgID, err := authenticator.Execute(r.Context(), key)
			if err != nil {
				apperror.Write(w, r, log, apperror.Unauthorized("invalid_api_key", "invalid or revoked API key"))
				return
			}
			ctx := authctx.With(r.Context(), authctx.Principal{OrgID: orgID})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireOwner rejects requests whose principal is not an owner (403).
func RequireOwner(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := authctx.From(r.Context())
			if !ok {
				apperror.Write(w, r, log, apperror.Unauthorized("unauthenticated", "authentication required"))
				return
			}
			if !p.IsOwner() {
				apperror.Write(w, r, log, apperror.Forbidden("owner_only", "this action is restricted to the owner"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}
