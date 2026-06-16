package domain

import "errors"

var (
	ErrExpenseIDRequired = errors.New("expense id is required")
	ErrOrgIDRequired     = errors.New("organization id is required")
	ErrCategoryRequired  = errors.New("expense category is required")
	ErrAmountNotPositive = errors.New("expense amount must be positive")
	ErrExpenseNotFound   = errors.New("expense not found")
)
