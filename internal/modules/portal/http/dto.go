package http

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// --- requests ---------------------------------------------------------------

type provisionAccessRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type sendMessageRequest struct {
	Body string `json:"body" validate:"required"`
}

// --- responses --------------------------------------------------------------

type loginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	ClientID  string    `json:"client_id"`
}

type accessResponse struct {
	HasAccess bool   `json:"has_access"`
	Email     string `json:"email"`
}

type messageResponse struct {
	ID         string    `json:"id"`
	SenderKind string    `json:"sender_kind"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

type conversationResponse struct {
	ClientID      string    `json:"client_id"`
	ClientName    string    `json:"client_name"`
	LastMessage   string    `json:"last_message"`
	LastSender    string    `json:"last_sender"`
	LastMessageAt time.Time `json:"last_message_at"`
}

type clientProfileResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type contractViewResponse struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	SalePrice    string    `json:"sale_price"`
	Outstanding  string    `json:"outstanding"`
	Status       string    `json:"status"`
	Installments int       `json:"installments"`
	CreatedAt    time.Time `json:"created_at"`
}

func toMessageResponse(m domain.Message) messageResponse {
	return messageResponse{
		ID:         m.ID(),
		SenderKind: m.SenderKind().String(),
		Body:       m.Body(),
		CreatedAt:  m.CreatedAt(),
	}
}

func toConversationResponse(v domain.ConversationView, name string) conversationResponse {
	return conversationResponse{
		ClientID:      v.ClientID,
		ClientName:    name,
		LastMessage:   v.LastMessage,
		LastSender:    v.LastSenderKind.String(),
		LastMessageAt: v.LastMessageAt,
	}
}

func toContractViewResponse(c domain.ContractView) contractViewResponse {
	return contractViewResponse{
		ID:           c.ID,
		ProductID:    c.ProductID,
		SalePrice:    c.SalePrice.String(),
		Outstanding:  c.Outstanding.String(),
		Status:       c.Status,
		Installments: c.Installments,
		CreatedAt:    c.CreatedAt,
	}
}
