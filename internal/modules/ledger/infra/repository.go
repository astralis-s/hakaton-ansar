// Package infra provides the ledger persistence adapter (sqlc repository) and the
// cross-context sales reader over financing.
package infra

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// ExpenseRepository implements domain.ExpenseRepository over sqlc.
type ExpenseRepository struct{ pool *pgxpool.Pool }

func NewExpenseRepository(pool *pgxpool.Pool) *ExpenseRepository {
	return &ExpenseRepository{pool: pool}
}

var _ domain.ExpenseRepository = (*ExpenseRepository)(nil)

func (r *ExpenseRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *ExpenseRepository) Create(ctx context.Context, e domain.Expense) (domain.Expense, error) {
	id, err := pgconv.UUID(e.ID())
	if err != nil {
		return domain.Expense{}, fmt.Errorf("invalid expense id: %w", err)
	}
	org, err := pgconv.UUID(e.OrgID())
	if err != nil {
		return domain.Expense{}, fmt.Errorf("invalid org id: %w", err)
	}
	amount, err := pgconv.Numeric(e.Amount().Amount())
	if err != nil {
		return domain.Expense{}, err
	}
	row, err := r.q(ctx).CreateExpense(ctx, sqlcgen.CreateExpenseParams{
		ID:       id,
		OrgID:    org,
		Category: e.Category(),
		Amount:   amount,
		Note:     e.Note(),
		SpentAt:  pgconv.Date(e.SpentAt()),
	})
	if err != nil {
		return domain.Expense{}, fmt.Errorf("create expense: %w", err)
	}
	return expenseFromRow(row)
}

func (r *ExpenseRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.Expense, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.q(ctx).ListExpensesByOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	out := make([]domain.Expense, 0, len(rows))
	for _, row := range rows {
		e, err := expenseFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *ExpenseRepository) Delete(ctx context.Context, orgID, id string) error {
	eid, err := pgconv.UUID(id)
	if err != nil {
		return domain.ErrExpenseNotFound
	}
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return domain.ErrExpenseNotFound
	}
	n, err := r.q(ctx).DeleteExpense(ctx, sqlcgen.DeleteExpenseParams{ID: eid, OrgID: org})
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}
	if n == 0 {
		return domain.ErrExpenseNotFound
	}
	return nil
}

func expenseFromRow(e sqlcgen.Expense) (domain.Expense, error) {
	amount, err := pgconv.DecimalFromNumeric(e.Amount)
	if err != nil {
		return domain.Expense{}, fmt.Errorf("decode amount: %w", err)
	}
	return domain.RehydrateExpense(
		pgconv.StrUUID(e.ID),
		pgconv.StrUUID(e.OrgID),
		e.Category,
		money.New(amount, money.DefaultCurrency),
		e.Note,
		pgconv.DateValue(e.SpentAt),
		pgconv.TimeValue(e.CreatedAt),
	), nil
}
