package domain

import (
	"strings"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Product is a catalog item available for installment sale. CostPrice is the
// purchase price; HalalStatus is mandatory and gates whether a contract may be
// created for it (a Haram product can be catalogued but not financed).
type Product struct {
	id          string
	orgID       string
	name        string
	category    string
	costPrice   money.Money
	halalStatus HalalStatus
	createdAt   time.Time
}

// NewProduct validates invariants and creates a fresh product.
func NewProduct(id, orgID, name, category string, costPrice money.Money, status HalalStatus) (Product, error) {
	if id == "" {
		return Product{}, ErrProductIDRequired
	}
	if orgID == "" {
		return Product{}, ErrOrgIDRequired
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Product{}, ErrProductNameRequired
	}
	if !costPrice.IsPositive() {
		return Product{}, ErrCostPriceNotPositive
	}
	if !status.Valid() {
		return Product{}, ErrInvalidHalalStatus
	}
	return Product{
		id:          id,
		orgID:       orgID,
		name:        name,
		category:    strings.TrimSpace(category),
		costPrice:   costPrice,
		halalStatus: status,
		createdAt:   time.Now().UTC(),
	}, nil
}

// RehydrateProduct rebuilds a product from trusted storage.
func RehydrateProduct(id, orgID, name, category string, costPrice money.Money, status HalalStatus, createdAt time.Time) Product {
	return Product{
		id:          id,
		orgID:       orgID,
		name:        name,
		category:    category,
		costPrice:   costPrice,
		halalStatus: status,
		createdAt:   createdAt,
	}
}

// Update returns a validated copy with new mutable fields, preserving identity
// and creation time. State changes go through methods, not field setters.
func (p Product) Update(name, category string, costPrice money.Money, status HalalStatus) (Product, error) {
	updated, err := NewProduct(p.id, p.orgID, name, category, costPrice, status)
	if err != nil {
		return Product{}, err
	}
	updated.createdAt = p.createdAt
	return updated, nil
}

func (p Product) ID() string               { return p.id }
func (p Product) OrgID() string            { return p.orgID }
func (p Product) Name() string             { return p.name }
func (p Product) Category() string         { return p.category }
func (p Product) CostPrice() money.Money   { return p.costPrice }
func (p Product) HalalStatus() HalalStatus { return p.halalStatus }
func (p Product) CreatedAt() time.Time     { return p.createdAt }

// IsHaram reports whether the product is forbidden.
func (p Product) IsHaram() bool { return p.halalStatus == HalalStatusHaram }

// CanBeFinanced reports whether a contract may be created for this product
// (everything except Haram).
func (p Product) CanBeFinanced() bool { return !p.IsHaram() }
