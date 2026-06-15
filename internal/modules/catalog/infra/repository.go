// Package infra provides the catalog persistence adapter (sqlc repository).
package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// ProductRepository implements domain.ProductRepository over sqlc.
type ProductRepository struct{ pool *pgxpool.Pool }

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

var _ domain.ProductRepository = (*ProductRepository)(nil)

func (r *ProductRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *ProductRepository) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	id, err := pgconv.UUID(p.ID())
	if err != nil {
		return domain.Product{}, fmt.Errorf("invalid product id: %w", err)
	}
	orgID, err := pgconv.UUID(p.OrgID())
	if err != nil {
		return domain.Product{}, fmt.Errorf("invalid org id: %w", err)
	}
	cost, err := pgconv.Numeric(p.CostPrice().Amount())
	if err != nil {
		return domain.Product{}, err
	}
	row, err := r.queries(ctx).CreateProduct(ctx, sqlcgen.CreateProductParams{
		ID:          id,
		OrgID:       orgID,
		Name:        p.Name(),
		Category:    p.Category(),
		CostPrice:   cost,
		HalalStatus: p.HalalStatus().String(),
	})
	if err != nil {
		return domain.Product{}, fmt.Errorf("create product: %w", err)
	}
	return productFromRow(row)
}

func (r *ProductRepository) GetByID(ctx context.Context, orgID, id string) (domain.Product, error) {
	pid, err := pgconv.UUID(id)
	if err != nil {
		return domain.Product{}, domain.ErrProductNotFound
	}
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return domain.Product{}, domain.ErrProductNotFound
	}
	row, err := r.queries(ctx).GetProductByID(ctx, sqlcgen.GetProductByIDParams{ID: pid, OrgID: org})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}
		return domain.Product{}, fmt.Errorf("get product: %w", err)
	}
	return productFromRow(row)
}

func (r *ProductRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.Product, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.queries(ctx).ListProductsByOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	products := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		p, err := productFromRow(row)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, p domain.Product) (domain.Product, error) {
	id, err := pgconv.UUID(p.ID())
	if err != nil {
		return domain.Product{}, domain.ErrProductNotFound
	}
	org, err := pgconv.UUID(p.OrgID())
	if err != nil {
		return domain.Product{}, domain.ErrProductNotFound
	}
	cost, err := pgconv.Numeric(p.CostPrice().Amount())
	if err != nil {
		return domain.Product{}, err
	}
	row, err := r.queries(ctx).UpdateProduct(ctx, sqlcgen.UpdateProductParams{
		ID:          id,
		OrgID:       org,
		Name:        p.Name(),
		Category:    p.Category(),
		CostPrice:   cost,
		HalalStatus: p.HalalStatus().String(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}
		return domain.Product{}, fmt.Errorf("update product: %w", err)
	}
	return productFromRow(row)
}

func productFromRow(p sqlcgen.Product) (domain.Product, error) {
	cost, err := pgconv.DecimalFromNumeric(p.CostPrice)
	if err != nil {
		return domain.Product{}, fmt.Errorf("decode cost price: %w", err)
	}
	return domain.RehydrateProduct(
		pgconv.StrUUID(p.ID),
		pgconv.StrUUID(p.OrgID),
		p.Name,
		p.Category,
		money.New(cost, money.DefaultCurrency),
		domain.HalalStatus(p.HalalStatus),
		pgconv.TimeValue(p.CreatedAt),
	), nil
}
