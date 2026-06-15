// Package authctx carries the authenticated principal across the request context
// in a transport- and module-agnostic way. The iam JWT middleware populates it;
// any module's handler reads it without importing iam's HTTP layer.
package authctx

import "context"

// Roles are kept as plain strings here to avoid coupling platform to the iam
// domain. The canonical values are "owner" and "manager".
const (
	RoleOwner   = "owner"
	RoleManager = "manager"
)

// Principal is the authenticated identity attached to a request.
type Principal struct {
	UserID string
	OrgID  string
	Role   string
}

func (p Principal) IsOwner() bool { return p.Role == RoleOwner }

type ctxKey struct{}

// With returns a copy of ctx carrying the principal.
func With(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// From extracts the principal from ctx, if present.
func From(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}
