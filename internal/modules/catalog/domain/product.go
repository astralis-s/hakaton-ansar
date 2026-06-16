package domain

import (
	"strings"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Product is a catalog item available for installment sale. CostPrice is the
// purchase price; HalalStatus is mandatory and gates whether a contract may be
// created for it (a Haram product can be catalogued but not financed). Stock is
// the on-hand quantity (товарооборот); a product with zero stock cannot be sold.
type Product struct {
	id          string
	orgID       string
	name        string
	category    string
	costPrice   money.Money
	halalStatus HalalStatus
	stock       int
	createdAt   time.Time
}

// NewProduct validates invariants and creates a fresh product with an initial
// stock (>= 0).
func NewProduct(id, orgID, name, category string, costPrice money.Money, status HalalStatus, stock int) (Product, error) {
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
	if stock < 0 {
		return Product{}, ErrNegativeStock
	}
	return Product{
		id:          id,
		orgID:       orgID,
		name:        name,
		category:    strings.TrimSpace(category),
		costPrice:   costPrice,
		halalStatus: status,
		stock:       stock,
		createdAt:   time.Now().UTC(),
	}, nil
}

// RehydrateProduct rebuilds a product from trusted storage.
func RehydrateProduct(id, orgID, name, category string, costPrice money.Money, status HalalStatus, stock int, createdAt time.Time) Product {
	return Product{
		id:          id,
		orgID:       orgID,
		name:        name,
		category:    category,
		costPrice:   costPrice,
		halalStatus: status,
		stock:       stock,
		createdAt:   createdAt,
	}
}

// Update returns a validated copy with new mutable fields, preserving identity,
// creation time and current stock (stock changes go through movements).
func (p Product) Update(name, category string, costPrice money.Money, status HalalStatus) (Product, error) {
	updated, err := NewProduct(p.id, p.orgID, name, category, costPrice, status, p.stock)
	if err != nil {
		return Product{}, err
	}
	updated.createdAt = p.createdAt
	return updated, nil
}

// WithStockDelta returns a copy with stock adjusted by delta. The resulting stock
// must not go negative (you cannot sell or write off more than is on hand).
func (p Product) WithStockDelta(delta int) (Product, error) {
	next := p.stock + delta
	if next < 0 {
		return Product{}, ErrInsufficientStock
	}
	updated := p
	updated.stock = next
	return updated, nil
}

func (p Product) ID() string               { return p.id }
func (p Product) OrgID() string            { return p.orgID }
func (p Product) Name() string             { return p.name }
func (p Product) Category() string         { return p.category }
func (p Product) CostPrice() money.Money   { return p.costPrice }
func (p Product) HalalStatus() HalalStatus { return p.halalStatus }
func (p Product) Stock() int               { return p.stock }
func (p Product) CreatedAt() time.Time     { return p.createdAt }

// InStock reports whether at least one unit is available.
func (p Product) InStock() bool { return p.stock > 0 }

// IsHaram reports whether the product is forbidden.
func (p Product) IsHaram() bool { return p.halalStatus == HalalStatusHaram }

// CanBeFinanced reports whether a contract may be created for this product
// (not Haram and in stock).
func (p Product) CanBeFinanced() bool { return !p.IsHaram() && p.InStock() }
