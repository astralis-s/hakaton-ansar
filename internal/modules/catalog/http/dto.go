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
}

type updateProductRequest struct {
	Name        string `json:"name" validate:"required"`
	Category    string `json:"category"`
	CostPrice   string `json:"cost_price" validate:"required"`
	HalalStatus string `json:"halal_status" validate:"required,oneof=halal haram doubtful"`
}

type productResponse struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	CostPrice     string    `json:"cost_price"` // decimal string
	HalalStatus   string    `json:"halal_status"`
	CanBeFinanced bool      `json:"can_be_financed"`
	CreatedAt     time.Time `json:"created_at"`
}

func toProductResponse(p domain.Product) productResponse {
	return productResponse{
		ID:            p.ID(),
		OrgID:         p.OrgID(),
		Name:          p.Name(),
		Category:      p.Category(),
		CostPrice:     p.CostPrice().String(),
		HalalStatus:   p.HalalStatus().String(),
		CanBeFinanced: p.CanBeFinanced(),
		CreatedAt:     p.CreatedAt(),
	}
}
