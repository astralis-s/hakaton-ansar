package domain

import "errors"

var (
	ErrProductIDRequired    = errors.New("product id is required")
	ErrOrgIDRequired        = errors.New("organization id is required")
	ErrProductNameRequired  = errors.New("product name is required")
	ErrCostPriceNotPositive = errors.New("cost price must be positive")
	ErrInvalidHalalStatus   = errors.New("invalid halal status: must be halal, haram or doubtful")
	ErrProductNotFound      = errors.New("product not found")

	ErrNegativeStock     = errors.New("stock must not be negative")
	ErrInsufficientStock = errors.New("not enough stock on hand")
	ErrStockDeltaZero    = errors.New("stock change must not be zero")
	ErrInvalidStockReason = errors.New("invalid stock reason: must be receipt, sale, adjustment or writeoff")
)
