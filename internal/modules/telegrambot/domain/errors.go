package domain

import "errors"

var (
	// ErrUserNotFound is returned by UserRepository when no row exists for a chat.
	ErrUserNotFound = errors.New("telegram user not found")

	ErrChatIDRequired   = errors.New("telegram chat id is required")
	ErrOrgIDRequired    = errors.New("organization id is required")
	ErrClientIDRequired = errors.New("client id is required")
	ErrInvalidState     = errors.New("action not allowed in the current registration state")
	ErrInvalidFullName  = errors.New("full name is invalid (need at least surname and name)")
	ErrInvalidPhone     = errors.New("phone is invalid")
)
