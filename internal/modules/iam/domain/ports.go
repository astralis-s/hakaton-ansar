package domain

import (
	"context"
	"time"
)

// OrganizationRepository persists organizations.
type OrganizationRepository interface {
	Create(ctx context.Context, org Organization) (Organization, error)
	GetByID(ctx context.Context, id string) (Organization, error)
	Count(ctx context.Context) (int64, error)
}

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, u User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	ListByOrg(ctx context.Context, orgID string) ([]User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// ApiKeyRepository persists public-API keys.
type ApiKeyRepository interface {
	Create(ctx context.Context, key ApiKey) (ApiKey, error)
	GetByHash(ctx context.Context, keyHash string) (ApiKey, error)
	ListByOrg(ctx context.Context, orgID string) ([]ApiKey, error)
	Revoke(ctx context.Context, id, orgID string) (ApiKey, error)
}

// PasswordHasher hashes and verifies user passwords. Implemented in infra with
// bcrypt (kept out of the sterile domain).
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

// Principal is the authenticated identity carried by a JWT.
type Principal struct {
	UserID string
	OrgID  string
	Role   Role
}

// TokenService issues and parses JWTs for the internal API. Implemented in infra
// with golang-jwt (kept out of the domain).
type TokenService interface {
	Issue(u User) (token string, expiresAt time.Time, err error)
	Parse(token string) (Principal, error)
}

// TxManager runs a function inside a single database transaction. The provided
// context carries the transaction so repositories transparently enlist in it.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
