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
	// SaveState persists a mutated outstanding balance and status.
	SaveState(ctx context.Context, c *Contract) error
	// AddPayment appends one payment row.
	AddPayment(ctx context.Context, contractID string, p Payment) error
}

// CharityRepository persists the sadaqa registry.
type CharityRepository interface {
	Create(ctx context.Context, e CharityEntry) (CharityEntry, error)
	ListByOrg(ctx context.Context, orgID string) ([]CharityEntry, error)
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

// ClientReader checks client existence in the crm context.
type ClientReader interface {
	Exists(ctx context.Context, orgID, clientID string) (bool, error)
}

// TxManager runs a function inside a single database transaction (the context
// carries the transaction so repositories enlist transparently).
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
