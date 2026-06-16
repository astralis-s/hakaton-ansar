package domain

import (
	"context"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// ContractSummary is a lightweight read model for contract lists (no schedule).
type ContractSummary struct {
	ID                string
	ClientID          string
	ProductID         string
	SalePrice         money.Money
	FinancedAmount    money.Money
	Outstanding       money.Money
	Status            ContractStatus
	InstallmentsCount int
	CreatedAt         time.Time
}

// ContractRepository persists contracts (aggregate root incl. schedule & payments).
type ContractRepository interface {
	Create(ctx context.Context, c *Contract) error
	GetByID(ctx context.Context, orgID, id string) (*Contract, error)
	ListByOrg(ctx context.Context, orgID string) ([]ContractSummary, error)
	// ListFullByOrg loads full aggregates (schedule + payments) for the org,
	// used by the dashboard aggregation.
	ListFullByOrg(ctx context.Context, orgID string) ([]*Contract, error)
	// SaveState persists a mutated outstanding balance and status.
	SaveState(ctx context.Context, c *Contract) error
	// AddPayment appends one payment row.
	AddPayment(ctx context.Context, contractID string, p Payment) error
}

// ProductInfo is the minimal product data financing needs to gate contract
// creation (it must refuse haram products).
type ProductInfo struct {
	ID      string
	IsHaram bool
}

// ProductReader reads product data from the catalog context (returns
// ErrProductNotFound when absent).
type ProductReader interface {
	Get(ctx context.Context, orgID, productID string) (ProductInfo, error)
}

// ClientReader checks client existence and resolves display names in the crm
// context.
type ClientReader interface {
	Exists(ctx context.Context, orgID, clientID string) (bool, error)
	// Names resolves client ids to their full names (id → name).
	Names(ctx context.Context, orgID string, ids []string) (map[string]string, error)
}

// StockReserver reserves one unit of a product when a contract is created. It
// runs inside the contract's transaction (so the stock write and the contract
// insert commit or roll back together) and returns ErrOutOfStock when no unit is
// available. Implemented in infra over the catalog stock repository.
type StockReserver interface {
	Reserve(ctx context.Context, orgID, productID string) error
}

// ContractRequestRepository persists client contract requests (заявки),
// org-scoped.
type ContractRequestRepository interface {
	Create(ctx context.Context, r *ContractRequest) error
	GetByID(ctx context.Context, orgID, id string) (*ContractRequest, error)
	ListByOrg(ctx context.Context, orgID string) ([]*ContractRequest, error)
	ListByClient(ctx context.Context, orgID, clientID string) ([]*ContractRequest, error)
	// Save persists a status transition (approve/reject) of an existing request.
	Save(ctx context.Context, r *ContractRequest) error
}

// TxManager runs a function inside a single database transaction (the context
// carries the transaction so repositories enlist transparently).
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
