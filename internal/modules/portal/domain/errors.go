package domain

import "errors"

var (
	ErrClientIDRequired       = errors.New("client id is required")
	ErrOrgIDRequired          = errors.New("organization id is required")
	ErrConversationIDRequired = errors.New("conversation id is required")
	ErrMessageIDRequired      = errors.New("message id is required")
	ErrSenderIDRequired       = errors.New("sender id is required")

	ErrInvalidEmail      = errors.New("invalid email")
	ErrEmailTaken        = errors.New("email is already in use")
	ErrPasswordRequired  = errors.New("password hash is required")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountNotFound   = errors.New("portal account not found")

	ErrInvalidSenderKind  = errors.New("invalid sender kind: must be client or staff")
	ErrMessageBodyRequired = errors.New("message body is required")
	ErrMessageTooLong     = errors.New("message body is too long")

	ErrClientNotFound   = errors.New("client not found")
	ErrContractNotFound = errors.New("contract not found")
	ErrProductNotFound  = errors.New("product not found")
	ErrProductHaram     = errors.New("product cannot be financed")
	ErrInvalidRequest   = errors.New("invalid contract request")
)
