package http

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/crm/domain"
)

type createClientRequest struct {
	FullName string `json:"full_name" validate:"required"`
	Phone    string `json:"phone"`
	Document string `json:"document"`
}

type updateClientRequest struct {
	FullName string `json:"full_name" validate:"required"`
	Phone    string `json:"phone"`
	Document string `json:"document"`
}

type clientResponse struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Document  string    `json:"document"`
	CreatedAt time.Time `json:"created_at"`
}

func toClientResponse(c domain.Client) clientResponse {
	return clientResponse{
		ID:        c.ID(),
		OrgID:     c.OrgID(),
		FullName:  c.FullName(),
		Phone:     c.Phone(),
		Document:  c.Document(),
		CreatedAt: c.CreatedAt(),
	}
}
