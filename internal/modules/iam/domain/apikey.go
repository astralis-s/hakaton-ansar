package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// ApiKey authorizes the PUBLIC API (/api/v1, header X-API-Key). It is bound to an
// organization and revocable. The raw secret is shown to the user exactly once;
// only its hash is persisted.
type ApiKey struct {
	id        string
	orgID     string
	name      string
	prefix    string // human-visible identifier (first chars of the raw key)
	keyHash   string // sha256(raw key), hex
	createdAt time.Time
	revokedAt *time.Time
}

// NewApiKey validates and creates a fresh, active API key record.
func NewApiKey(id, orgID, name, prefix, keyHash string) (ApiKey, error) {
	if id == "" {
		return ApiKey{}, ErrUserIDRequired
	}
	if orgID == "" {
		return ApiKey{}, ErrOrgIDRequired
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ApiKey{}, ErrApiKeyNameRequired
	}
	if keyHash == "" || prefix == "" {
		return ApiKey{}, ErrApiKeyNameRequired
	}
	return ApiKey{
		id:        id,
		orgID:     orgID,
		name:      name,
		prefix:    prefix,
		keyHash:   keyHash,
		createdAt: time.Now().UTC(),
	}, nil
}

// RehydrateApiKey rebuilds an API key from trusted storage.
func RehydrateApiKey(id, orgID, name, prefix, keyHash string, createdAt time.Time, revokedAt *time.Time) ApiKey {
	return ApiKey{
		id:        id,
		orgID:     orgID,
		name:      name,
		prefix:    prefix,
		keyHash:   keyHash,
		createdAt: createdAt,
		revokedAt: revokedAt,
	}
}

func (k ApiKey) ID() string            { return k.id }
func (k ApiKey) OrgID() string         { return k.orgID }
func (k ApiKey) Name() string          { return k.name }
func (k ApiKey) Prefix() string        { return k.prefix }
func (k ApiKey) KeyHash() string       { return k.keyHash }
func (k ApiKey) CreatedAt() time.Time  { return k.createdAt }
func (k ApiKey) RevokedAt() *time.Time { return k.revokedAt }
func (k ApiKey) IsActive() bool        { return k.revokedAt == nil }

// HashAPIKey hashes a raw API key deterministically. API keys are high-entropy,
// so a fast cryptographic hash (sha256) is appropriate — unlike user passwords,
// which use bcrypt. Used both on creation and on verification.
func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
