package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
)

// StockRepository implements domain.StockRepository over sqlc. Adjust must run
// inside a transaction (it locks the product row FOR UPDATE, then writes the new
// balance and the movement); callers provide the transaction via the context.
type StockRepository struct{ pool *pgxpool.Pool }

func NewStockRepository(pool *pgxpool.Pool) *StockRepository {
	return &StockRepository{pool: pool}
}

var _ domain.StockRepository = (*StockRepository)(nil)

func (r *StockRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *StockRepository) Adjust(ctx context.Context, orgID, productID string, delta int, reason domain.StockReason, note string) (domain.Product, domain.StockMovement, error) {
	pid, err := pgconv.UUID(productID)
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, domain.ErrProductNotFound
	}
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, domain.ErrProductNotFound
	}

	row, err := r.q(ctx).GetProductForUpdate(ctx, sqlcgen.GetProductForUpdateParams{ID: pid, OrgID: org})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.StockMovement{}, domain.ErrProductNotFound
		}
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("lock product: %w", err)
	}
	product, err := productFromRow(row)
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, err
	}

	updated, err := product.WithStockDelta(delta)
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, err // ErrInsufficientStock
	}

	if _, err := r.q(ctx).SetProductStock(ctx, sqlcgen.SetProductStockParams{
		ID: pid, OrgID: org, Stock: int32(updated.Stock()),
	}); err != nil {
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("set stock: %w", err)
	}

	movement, err := domain.NewStockMovement(uuid.NewString(), orgID, productID, delta, reason, note, updated.Stock())
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, err
	}
	mid, _ := pgconv.UUID(movement.ID())
	mRow, err := r.q(ctx).CreateStockMovement(ctx, sqlcgen.CreateStockMovementParams{
		ID:           mid,
		OrgID:        org,
		ProductID:    pid,
		Delta:        int32(delta),
		Reason:       reason.String(),
		Note:         note,
		BalanceAfter: int32(updated.Stock()),
	})
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("create stock movement: %w", err)
	}
	return updated, movementFromRow(mRow), nil
}

func (r *StockRepository) ListMovementsByOrg(ctx context.Context, orgID string) ([]domain.StockMovement, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.q(ctx).ListStockMovementsByOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list stock movements: %w", err)
	}
	out := make([]domain.StockMovement, 0, len(rows))
	for _, row := range rows {
		out = append(out, movementFromRow(row))
	}
	return out, nil
}

func movementFromRow(m sqlcgen.StockMovement) domain.StockMovement {
	return domain.RehydrateStockMovement(
		pgconv.StrUUID(m.ID),
		pgconv.StrUUID(m.OrgID),
		pgconv.StrUUID(m.ProductID),
		int(m.Delta),
		domain.StockReason(m.Reason),
		m.Note,
		int(m.BalanceAfter),
		pgconv.TimeValue(m.CreatedAt),
	)
}
