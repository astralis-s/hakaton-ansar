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

// StockRepository manages product stock balances and their movement log
// (товарооборот). Adjust changes a product's stock by delta and records a
// movement atomically; it returns ErrInsufficientStock if the balance would go
// negative. Used both by the catalog (manual receipts/write-offs) and by
// financing (reserve one unit on sale, in the same transaction as the contract).
type StockRepository interface {
	Adjust(ctx context.Context, orgID, productID string, delta int, reason StockReason, note string) (Product, StockMovement, error)
	ListMovementsByOrg(ctx context.Context, orgID string) ([]StockMovement, error)
}

// TxManager runs a function inside a single database transaction (the context
// carries the transaction so repositories enlist transparently).
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
