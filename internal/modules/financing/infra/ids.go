package infra

import "github.com/google/uuid"

// newID generates an identifier for child rows (installments) created during a
// contract insert.
func newID() string { return uuid.NewString() }
