// Package app holds the ledger use-cases (one type per scenario).
package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// CreateExpense records a manual business expense.
type CreateExpense struct {
	expenses domain.ExpenseRepository
}

func NewCreateExpense(expenses domain.ExpenseRepository) *CreateExpense {
	return &CreateExpense{expenses: expenses}
}

type CreateExpenseInput struct {
	OrgID    string
	Category string
	Amount   money.Money
	Note     string
	SpentAt  time.Time
}

func (uc *CreateExpense) Execute(ctx context.Context, in CreateExpenseInput) (domain.Expense, error) {
	expense, err := domain.NewExpense(uuid.NewString(), in.OrgID, in.Category, in.Amount, in.Note, in.SpentAt)
	if err != nil {
		return domain.Expense{}, err
	}
	return uc.expenses.Create(ctx, expense)
}

// ListExpenses returns the organization's manual expenses.
type ListExpenses struct {
	expenses domain.ExpenseRepository
}

func NewListExpenses(expenses domain.ExpenseRepository) *ListExpenses {
	return &ListExpenses{expenses: expenses}
}

func (uc *ListExpenses) Execute(ctx context.Context, orgID string) ([]domain.Expense, error) {
	return uc.expenses.ListByOrg(ctx, orgID)
}

// DeleteExpense removes a manual expense (corrections).
type DeleteExpense struct {
	expenses domain.ExpenseRepository
}

func NewDeleteExpense(expenses domain.ExpenseRepository) *DeleteExpense {
	return &DeleteExpense{expenses: expenses}
}

func (uc *DeleteExpense) Execute(ctx context.Context, orgID, id string) error {
	return uc.expenses.Delete(ctx, orgID, id)
}
