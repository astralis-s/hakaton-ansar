package domain

import "context"

// ProductRepository persists catalog products. All reads/writes are scoped to an
// organization for tenant isolation.
type ProductRepository interface {
	Create(ctx context.Context, p Product) (Product, error)
	GetByID(ctx context.Context, orgID, id string) (Product, error)
	ListByOrg(ctx context.Context, orgID string) ([]Product, error)
	Update(ctx context.Context, p Product) (Product, error)
}
