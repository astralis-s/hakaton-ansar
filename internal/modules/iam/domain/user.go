package domain

import (
	"strings"
	"time"
)

// User belongs to an organization and authenticates against the internal API
// (/api/app) by JWT. The password is never stored in plain text — only its hash
// (produced by a PasswordHasher port; bcrypt lives in infra, not the domain).
type User struct {
	id           string
	orgID        string
	fullName     string
	email        string
	passwordHash string
	role         Role
	createdAt    time.Time
}

// NewUser validates invariants and creates a fresh user. The caller supplies an
// already-hashed password (hashing is an infra concern).
func NewUser(id, orgID, fullName, email, passwordHash string, role Role) (User, error) {
	if id == "" {
		return User{}, ErrUserIDRequired
	}
	if orgID == "" {
		return User{}, ErrOrgIDRequired
	}
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return User{}, ErrFullNameRequired
	}
	normEmail, err := normalizeEmail(email)
	if err != nil {
		return User{}, err
	}
	if passwordHash == "" {
		return User{}, ErrPasswordHashEmpty
	}
	if !role.Valid() {
		return User{}, ErrInvalidRole
	}
	return User{
		id:           id,
		orgID:        orgID,
		fullName:     fullName,
		email:        normEmail,
		passwordHash: passwordHash,
		role:         role,
		createdAt:    time.Now().UTC(),
	}, nil
}

// RehydrateUser rebuilds a user from trusted storage (no re-validation).
func RehydrateUser(id, orgID, fullName, email, passwordHash string, role Role, createdAt time.Time) User {
	return User{
		id:           id,
		orgID:        orgID,
		fullName:     fullName,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		createdAt:    createdAt,
	}
}

func (u User) ID() string           { return u.id }
func (u User) OrgID() string        { return u.orgID }
func (u User) FullName() string     { return u.fullName }
func (u User) Email() string        { return u.email }
func (u User) PasswordHash() string { return u.passwordHash }
func (u User) Role() Role           { return u.role }
func (u User) CreatedAt() time.Time { return u.createdAt }
