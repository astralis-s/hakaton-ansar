// Package app holds the financing use-cases (one type per scenario).
package app

import "github.com/google/uuid"

// NewID generates a new UUID string for an entity identifier.
func NewID() string { return uuid.NewString() }
