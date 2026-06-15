// Package app holds the catalog use-cases (one type per scenario).
package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// CreateProduct adds a product to the catalog.
type CreateProduct struct {
	products domain.ProductRepository
}

func NewCreateProduct(products domain.ProductRepository) *CreateProduct {
	return &CreateProduct{products: products}
}

type CreateProductInput struct {
	OrgID       string
	Name        string
	Category    string
	CostPrice   money.Money
	HalalStatus string
}

func (uc *CreateProduct) Execute(ctx context.Context, in CreateProductInput) (domain.Product, error) {
	status, err := domain.ParseHalalStatus(in.HalalStatus)
	if err != nil {
		return domain.Product{}, err
	}
	product, err := domain.NewProduct(uuid.NewString(), in.OrgID, in.Name, in.Category, in.CostPrice, status)
	if err != nil {
		return domain.Product{}, err
	}
	return uc.products.Create(ctx, product)
}

// GetProduct returns a single product within an organization.
type GetProduct struct {
	products domain.ProductRepository
}

func NewGetProduct(products domain.ProductRepository) *GetProduct {
	return &GetProduct{products: products}
}

func (uc *GetProduct) Execute(ctx context.Context, orgID, id string) (domain.Product, error) {
	return uc.products.GetByID(ctx, orgID, id)
}

// ListProducts returns all products of an organization.
type ListProducts struct {
	products domain.ProductRepository
}

func NewListProducts(products domain.ProductRepository) *ListProducts {
	return &ListProducts{products: products}
}

func (uc *ListProducts) Execute(ctx context.Context, orgID string) ([]domain.Product, error) {
	return uc.products.ListByOrg(ctx, orgID)
}

// UpdateProduct edits an existing product (identity and creation time preserved).
type UpdateProduct struct {
	products domain.ProductRepository
}

func NewUpdateProduct(products domain.ProductRepository) *UpdateProduct {
	return &UpdateProduct{products: products}
}

type UpdateProductInput struct {
	OrgID       string
	ID          string
	Name        string
	Category    string
	CostPrice   money.Money
	HalalStatus string
}

func (uc *UpdateProduct) Execute(ctx context.Context, in UpdateProductInput) (domain.Product, error) {
	status, err := domain.ParseHalalStatus(in.HalalStatus)
	if err != nil {
		return domain.Product{}, err
	}
	existing, err := uc.products.GetByID(ctx, in.OrgID, in.ID)
	if err != nil {
		return domain.Product{}, err
	}
	updated, err := existing.Update(in.Name, in.Category, in.CostPrice, status)
	if err != nil {
		return domain.Product{}, err
	}
	return uc.products.Update(ctx, updated)
}
