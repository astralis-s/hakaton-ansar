// Package domain is the client-portal bounded context: it lets a client log in
// (separate credentials from staff users) and chat with the company's staff. The
// domain depends only on stdlib (no infrastructure).
package domain

import (
	"strings"
	"time"
)

// PortalAccount is the login credential that lets a CRM client access the portal.
// It is provisioned by staff (one account per client). The password hash is
// computed in the infrastructure layer (bcrypt) — the domain only stores it.
type PortalAccount struct {
	clientID     string
	orgID        string
	email        string
	passwordHash string
	createdAt    time.Time
}

// NewPortalAccount validates invariants and creates an account.
func NewPortalAccount(clientID, orgID, email, passwordHash string) (PortalAccount, error) {
	if clientID == "" {
		return PortalAccount{}, ErrClientIDRequired
	}
	if orgID == "" {
		return PortalAccount{}, ErrOrgIDRequired
	}
	email = normalizeEmail(email)
	if email == "" || !strings.Contains(email, "@") {
		return PortalAccount{}, ErrInvalidEmail
	}
	if passwordHash == "" {
		return PortalAccount{}, ErrPasswordRequired
	}
	return PortalAccount{
		clientID:     clientID,
		orgID:        orgID,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    time.Now().UTC(),
	}, nil
}

// RehydratePortalAccount rebuilds an account from trusted storage.
func RehydratePortalAccount(clientID, orgID, email, passwordHash string, createdAt time.Time) PortalAccount {
	return PortalAccount{
		clientID:     clientID,
		orgID:        orgID,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    createdAt,
	}
}

// NormalizeEmail lowercases and trims an email for consistent lookups.
func NormalizeEmail(email string) string { return normalizeEmail(email) }

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (a PortalAccount) ClientID() string     { return a.clientID }
func (a PortalAccount) OrgID() string         { return a.orgID }
func (a PortalAccount) Email() string          { return a.email }
func (a PortalAccount) PasswordHash() string   { return a.passwordHash }
func (a PortalAccount) CreatedAt() time.Time  { return a.createdAt }
