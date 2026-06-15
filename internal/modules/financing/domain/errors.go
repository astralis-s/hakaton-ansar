package domain

import "errors"

var (
	// Construction / invariant errors.
	ErrContractIDRequired          = errors.New("contract id is required")
	ErrOrgIDRequired               = errors.New("organization id is required")
	ErrClientIDRequired            = errors.New("client id is required")
	ErrProductIDRequired           = errors.New("product id is required")
	ErrCostPriceNotPositive        = errors.New("cost price must be positive")
	ErrMarkupNegative              = errors.New("markup must not be negative")
	ErrDownPaymentNegative         = errors.New("down payment must not be negative")
	ErrDownPaymentTooLarge         = errors.New("down payment must be less than sale price")
	ErrInstallmentsNotPositive     = errors.New("installments must be at least 1")
	ErrInvalidCadence              = errors.New("invalid cadence: must be weekly or monthly")
	ErrFinancedLessThanInstallment = errors.New("financed amount must be at least the number of installments (each installment ≥ 1 minor unit)")
	ErrCurrencyMismatch            = errors.New("currency mismatch")
	ErrScheduleMismatch            = errors.New("schedule does not sum to the financed amount")

	// State-machine / operation errors.
	ErrContractNotActive         = errors.New("contract is not active")
	ErrInvalidStatusTransition   = errors.New("invalid contract status transition")
	ErrPaymentNotPositive        = errors.New("payment amount must be positive")
	ErrPaymentExceedsOutstanding = errors.New("payment exceeds outstanding balance")
	ErrAlreadySettled            = errors.New("contract is already settled")

	// Charity.
	ErrCharityAmountNotPositive = errors.New("charity amount must be positive")

	// Lookups / cross-module.
	ErrContractNotFound = errors.New("contract not found")
	ErrProductNotFound  = errors.New("product not found")
	ErrClientNotFound   = errors.New("client not found")
	ErrProductHaram     = errors.New("cannot create a contract for a haram product")
)
