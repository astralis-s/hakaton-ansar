package domain

import (
	"net/mail"
	"strings"
)

// Role is the privileged-role enumeration. Product language: «владелец / менеджер»
// (CHANGELOG canon #2). There is intentionally no "admin".
type Role string

const (
	RoleOwner   Role = "owner"
	RoleManager Role = "manager"
)

// ParseRole validates and normalizes a role string.
func ParseRole(s string) (Role, error) {
	switch Role(strings.ToLower(strings.TrimSpace(s))) {
	case RoleOwner:
		return RoleOwner, nil
	case RoleManager:
		return RoleManager, nil
	default:
		return "", ErrInvalidRole
	}
}

func (r Role) Valid() bool    { return r == RoleOwner || r == RoleManager }
func (r Role) IsOwner() bool  { return r == RoleOwner }
func (r Role) String() string { return string(r) }

// normalizeEmail lowercases and trims an email, validating its shape via the
// standard library (net/mail) — no external dependency, domain stays sterile.
func normalizeEmail(raw string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", ErrInvalidEmail
	}
	// ParseAddress accepts "Name <a@b>"; we only want the bare address.
	if addr.Address != trimmed {
		return "", ErrInvalidEmail
	}
	return trimmed, nil
}
