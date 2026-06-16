package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
)

type clientCtxKey struct{}

func withClient(ctx context.Context, p domain.ClientPrincipal) context.Context {
	return context.WithValue(ctx, clientCtxKey{}, p)
}

func clientFrom(ctx context.Context) (domain.ClientPrincipal, bool) {
	p, ok := ctx.Value(clientCtxKey{}).(domain.ClientPrincipal)
	return p, ok
}

// ClientAuth validates the client portal JWT ("Authorization: Bearer <jwt>" with
// kind=client) for the /api/portal surface and attaches the client principal.
func ClientAuth(tokens domain.ClientTokenService, log *slog.Logger) func(http.Handler) http.Handler {
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
			next.ServeHTTP(w, r.WithContext(withClient(r.Context(), principal)))
		})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}
