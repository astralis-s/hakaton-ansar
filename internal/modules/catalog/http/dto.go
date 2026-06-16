package http

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
)

type createProductRequest struct {
	Name        string `json:"name" validate:"required"`
	Category    string `json:"category"`
	CostPrice   string `json:"cost_price" validate:"required"` // decimal string, e.g. "85000.00"
	HalalStatus string `json:"halal_status" validate:"required,oneof=halal haram doubtful"`
	Stock       int    `json:"stock" validate:"min=0"` // initial on-hand quantity
}

type updateProductRequest struct {
	Name        string `json:"name" validate:"required"`
	Category    string `json:"category"`
	CostPrice   string `json:"cost_price" validate:"required"`
	HalalStatus string `json:"halal_status" validate:"required,oneof=halal haram doubtful"`
}

type adjustStockRequest struct {
	Delta  int    `json:"delta" validate:"required"` // +receipt / -writeoff / correction
	Reason string `json:"reason" validate:"required,oneof=receipt adjustment writeoff"`
	Note   string `json:"note"`
}

type productResponse struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	CostPrice     string    `json:"cost_price"` // decimal string
	HalalStatus   string    `json:"halal_status"`
	Stock         int       `json:"stock"`
	InStock       bool      `json:"in_stock"`
	CanBeFinanced bool      `json:"can_be_financed"`
	CreatedAt     time.Time `json:"created_at"`
}

type stockMovementResponse struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	Delta        int       `json:"delta"`
	Reason       string    `json:"reason"`
	Note         string    `json:"note"`
	BalanceAfter int       `json:"balance_after"`
	CreatedAt    time.Time `json:"created_at"`
}

func toProductResponse(p domain.Product) productResponse {
	return productResponse{
		ID:            p.ID(),
		OrgID:         p.OrgID(),
		Name:          p.Name(),
		Category:      p.Category(),
		CostPrice:     p.CostPrice().String(),
		HalalStatus:   p.HalalStatus().String(),
		Stock:         p.Stock(),
		InStock:       p.InStock(),
		CanBeFinanced: p.CanBeFinanced(),
		CreatedAt:     p.CreatedAt(),
	}
}

func toStockMovementResponse(m domain.StockMovement) stockMovementResponse {
	return stockMovementResponse{
		ID:           m.ID(),
		ProductID:    m.ProductID(),
		Delta:        m.Delta(),
		Reason:       m.Reason().String(),
		Note:         m.Note(),
		BalanceAfter: m.BalanceAfter(),
		CreatedAt:    m.CreatedAt(),
	}
}
