// Package domain is the ledger bounded context: income/expense accounting
// (учёт доходов и расходов). Income is derived from financing contracts
// (sale − purchase); expenses are the cost of goods sold (also derived) plus
// manually recorded business costs (rent, repairs, …). The domain depends only
// on stdlib + money (no infrastructure).
package domain

import (
	"strings"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Expense is a manually recorded business cost that is NOT the purchase price of
// a sold good (rent, repairs, logistics, salaries, …). The cost of goods sold is
// derived from contracts, not stored here — see Report / BuildReport.
type Expense struct {
	id        string
	orgID     string
	category  string
	amount    money.Money
	note      string
	spentAt   time.Time
	createdAt time.Time
}

// NewExpense validates invariants and creates a manual expense. A zero spentAt
// defaults to now (the expense is recorded as happening today).
func NewExpense(id, orgID, category string, amount money.Money, note string, spentAt time.Time) (Expense, error) {
	if id == "" {
		return Expense{}, ErrExpenseIDRequired
	}
	if orgID == "" {
		return Expense{}, ErrOrgIDRequired
	}
	category = strings.TrimSpace(category)
	if category == "" {
		return Expense{}, ErrCategoryRequired
	}
	if !amount.IsPositive() {
		return Expense{}, ErrAmountNotPositive
	}
	if spentAt.IsZero() {
		spentAt = time.Now()
	}
	return Expense{
		id:        id,
		orgID:     orgID,
		category:  category,
		amount:    amount,
		note:      strings.TrimSpace(note),
		spentAt:   spentAt.UTC(),
		createdAt: time.Now().UTC(),
	}, nil
}

// RehydrateExpense rebuilds an expense from trusted storage.
func RehydrateExpense(id, orgID, category string, amount money.Money, note string, spentAt, createdAt time.Time) Expense {
	return Expense{
		id:        id,
		orgID:     orgID,
		category:  category,
		amount:    amount,
		note:      note,
		spentAt:   spentAt,
		createdAt: createdAt,
	}
}

func (e Expense) ID() string           { return e.id }
func (e Expense) OrgID() string         { return e.orgID }
func (e Expense) Category() string      { return e.category }
func (e Expense) Amount() money.Money   { return e.amount }
func (e Expense) Note() string          { return e.note }
func (e Expense) SpentAt() time.Time    { return e.spentAt }
func (e Expense) CreatedAt() time.Time  { return e.createdAt }
