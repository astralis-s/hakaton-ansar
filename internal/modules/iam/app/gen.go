package app

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

// NewID generates a new UUID string for entity identifiers (the domain treats
// IDs as opaque strings; generation is an application concern).
func NewID() string {
	return uuid.NewString()
}

// generateAPIKey returns a fresh high-entropy public-API key and its display
// prefix. The raw value is returned to the caller exactly once and never stored.
func generateAPIKey() (raw, prefix string, err error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = "amana_" + base64.RawURLEncoding.EncodeToString(b)
	prefix = raw[:14] // "amana_" + 8 chars, safe to display
	return raw, prefix, nil
}
