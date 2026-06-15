package domain

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Cadence is the payment periodicity. It only affects due dates, never amounts.
type Cadence string

const (
	CadenceWeekly  Cadence = "weekly"
	CadenceMonthly Cadence = "monthly"
)

func ParseCadence(s string) (Cadence, error) {
	switch Cadence(strings.ToLower(strings.TrimSpace(s))) {
	case CadenceWeekly:
		return CadenceWeekly, nil
	case CadenceMonthly:
		return CadenceMonthly, nil
	default:
		return "", ErrInvalidCadence
	}
}

func (c Cadence) Valid() bool    { return c == CadenceWeekly || c == CadenceMonthly }
func (c Cadence) String() string { return string(c) }

// addN returns start shifted by n periods (n is 0-based for installment #n+1).
func (c Cadence) addN(start time.Time, n int) time.Time {
	switch c {
	case CadenceWeekly:
		return start.AddDate(0, 0, 7*n)
	default: // monthly
		return start.AddDate(0, n, 0)
	}
}

// PeriodsPerYear is used only for the illustrative riba comparison.
func (c Cadence) PeriodsPerYear() int {
	if c == CadenceWeekly {
		return 52
	}
	return 12
}

// ContractStatus is the aggregate's lifecycle state.
type ContractStatus string

const (
	StatusDraft     ContractStatus = "draft"
	StatusActive    ContractStatus = "active"
	StatusCompleted ContractStatus = "completed"
	StatusCancelled ContractStatus = "cancelled"
)

func (s ContractStatus) String() string { return string(s) }

// InstallmentStatus is derived from accumulated payment, never stored independently.
type InstallmentStatus string

const (
	InstallmentPending       InstallmentStatus = "pending"
	InstallmentPartiallyPaid InstallmentStatus = "partially_paid"
	InstallmentPaid          InstallmentStatus = "paid"
	InstallmentOverdue       InstallmentStatus = "overdue"
)

func (s InstallmentStatus) String() string { return string(s) }

// Markup is a fixed surcharge over the cost price. It is always >= 0, and once
// constructed it is a fixed amount — never a rate (anti-riba: the debt does not
// grow with time).
type Markup struct {
	amount money.Money
}

// NewMarkup wraps a money amount, requiring it to be non-negative.
func NewMarkup(amount money.Money) (Markup, error) {
	if amount.IsNegative() {
		return Markup{}, ErrMarkupNegative
	}
	return Markup{amount: amount}, nil
}

// NewMarkupFromPercent computes a markup as percent of the cost price and fixes
// it as an amount (it is not stored as a rate and is never recomputed).
func NewMarkupFromPercent(cost money.Money, percent decimal.Decimal) (Markup, error) {
	if percent.IsNegative() {
		return Markup{}, ErrMarkupNegative
	}
	factor := percent.Div(decimal.NewFromInt(100))
	return NewMarkup(cost.Mul(factor))
}

func (m Markup) Money() money.Money { return m.amount }
