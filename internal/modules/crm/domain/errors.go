package domain

import "errors"

var (
	ErrClientIDRequired   = errors.New("client id is required")
	ErrOrgIDRequired      = errors.New("organization id is required")
	ErrClientNameRequired = errors.New("client full name is required")
	ErrClientNotFound     = errors.New("client not found")
)
