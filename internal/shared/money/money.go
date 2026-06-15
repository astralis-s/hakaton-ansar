// Package money is the project's shared-kernel money value object. Several
// bounded contexts use it (catalog CostPrice, financing SalePrice/Markup/...),
// so it lives in a shared kernel rather than inside one module. It is a pure
// value object — it imports only the standard library and shopspring/decimal,
// so domain packages may depend on it without breaking their sterility rule.
//
// Rules (CLAUDE.md §0.1): money is never a float; the exact value is a decimal;
// at the API boundary it serializes as a decimal string ("120000.00"); internal
// schedule math uses int64 minor units (cents) via Cents().
package money

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
)

// DefaultCurrency is used when a currency is not specified (MVP is single-currency).
const DefaultCurrency = "RUB"

var (
	ErrInvalidAmount    = errors.New("invalid money amount")
	ErrCurrencyMismatch = errors.New("money currency mismatch")
)

// Money is an exact monetary amount in a currency, normalized to 2 decimal places.
type Money struct {
	amount   decimal.Decimal
	currency string
}

// New builds Money from a decimal, rounding to 2 places. Empty currency → default.
func New(amount decimal.Decimal, currency string) Money {
	if currency == "" {
		currency = DefaultCurrency
	}
	return Money{amount: amount.Round(2), currency: currency}
}

// FromString parses a decimal string (e.g. "120000.00") into Money.
func FromString(s, currency string) (Money, error) {
	d, err := decimal.NewFromString(strings.TrimSpace(s))
	if err != nil {
		return Money{}, ErrInvalidAmount
	}
	return New(d, currency), nil
}

// FromCents builds Money from int64 minor units.
func FromCents(cents int64, currency string) Money {
	return New(decimal.New(cents, -2), currency)
}

// Zero is a zero amount in the given currency.
func Zero(currency string) Money {
	return New(decimal.Zero, currency)
}

func (m Money) Amount() decimal.Decimal { return m.amount }
func (m Money) Currency() string        { return m.currency }

// String renders the amount with exactly 2 decimals — the API representation.
func (m Money) String() string { return m.amount.StringFixed(2) }

// Cents returns the amount in int64 minor units (for deterministic schedule math).
func (m Money) Cents() int64 {
	return m.amount.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
}

func (m Money) IsZero() bool     { return m.amount.IsZero() }
func (m Money) IsPositive() bool { return m.amount.IsPositive() }
func (m Money) IsNegative() bool { return m.amount.IsNegative() }

// Equals reports value+currency equality.
func (m Money) Equals(o Money) bool {
	return m.currency == o.currency && m.amount.Equal(o.amount)
}

// Cmp compares two same-currency amounts: -1, 0, or 1. Different currencies → error.
func (m Money) Cmp(o Money) (int, error) {
	if m.currency != o.currency {
		return 0, ErrCurrencyMismatch
	}
	return m.amount.Cmp(o.amount), nil
}

// Add returns m+o (same currency required).
func (m Money) Add(o Money) (Money, error) {
	if m.currency != o.currency {
		return Money{}, ErrCurrencyMismatch
	}
	return New(m.amount.Add(o.amount), m.currency), nil
}

// Sub returns m-o (same currency required).
func (m Money) Sub(o Money) (Money, error) {
	if m.currency != o.currency {
		return Money{}, ErrCurrencyMismatch
	}
	return New(m.amount.Sub(o.amount), m.currency), nil
}
