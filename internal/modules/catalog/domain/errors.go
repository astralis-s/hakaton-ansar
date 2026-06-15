package domain

import "errors"

var (
	ErrProductIDRequired    = errors.New("product id is required")
	ErrOrgIDRequired        = errors.New("organization id is required")
	ErrProductNameRequired  = errors.New("product name is required")
	ErrCostPriceNotPositive = errors.New("cost price must be positive")
	ErrInvalidHalalStatus   = errors.New("invalid halal status: must be halal, haram or doubtful")
	ErrProductNotFound      = errors.New("product not found")
)
