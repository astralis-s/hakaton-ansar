package domain

import "errors"

var (
	ErrOrgNameRequired    = errors.New("organization name is required")
	ErrUserIDRequired     = errors.New("user id is required")
	ErrOrgIDRequired      = errors.New("organization id is required")
	ErrFullNameRequired   = errors.New("full name is required")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrPasswordHashEmpty  = errors.New("password hash is required")
	ErrInvalidRole        = errors.New("invalid role: must be owner or manager")
	ErrApiKeyNameRequired = errors.New("api key name is required")

	ErrUserNotFound       = errors.New("user not found")
	ErrOrgNotFound        = errors.New("organization not found")
	ErrApiKeyNotFound     = errors.New("api key not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already in use")
	ErrAlreadyInitialized = errors.New("organization already initialized")
)
