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

type submitRequestRequest struct {
	ProductID           string `json:"product_id" validate:"required,uuid"`
	DesiredInstallments int    `json:"desired_installments" validate:"required,min=1"`
	DesiredDownPayment  string `json:"desired_down_payment"`
	Note                string `json:"note"`
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

type telegramLinkResponse struct {
	Available bool   `json:"available"`
	URL       string `json:"url,omitempty"`
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

const dateLayout = "2006-01-02"

type installmentLineResponse struct {
	Number  int    `json:"number"`
	DueDate string `json:"due_date"`
	Amount  string `json:"amount"`
	Status  string `json:"status"`
}

type paymentLineResponse struct {
	Amount string    `json:"amount"`
	PaidAt time.Time `json:"paid_at"`
}

type contractDetailResponse struct {
	ID             string                    `json:"id"`
	ProductID      string                    `json:"product_id"`
	SalePrice      string                    `json:"sale_price"`
	DownPayment    string                    `json:"down_payment"`
	FinancedAmount string                    `json:"financed_amount"`
	Outstanding    string                    `json:"outstanding"`
	PaidAmount     string                    `json:"paid_amount"`
	Status         string                    `json:"status"`
	Cadence        string                    `json:"cadence"`
	StartDate      string                    `json:"start_date"`
	HasOverdue     bool                      `json:"has_overdue"`
	HasNext        bool                      `json:"has_next"`
	NextDueDate    string                    `json:"next_due_date"`
	NextDueAmount  string                    `json:"next_due_amount"`
	Schedule       []installmentLineResponse `json:"schedule"`
	Payments       []paymentLineResponse     `json:"payments"`
	CreatedAt      time.Time                 `json:"created_at"`
}

type productCardResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type requestViewResponse struct {
	ID                  string    `json:"id"`
	ProductID           string    `json:"product_id"`
	DesiredInstallments int       `json:"desired_installments"`
	DesiredDownPayment  string    `json:"desired_down_payment"`
	Note                string    `json:"note"`
	Status              string    `json:"status"`
	ContractID          string    `json:"contract_id"`
	CreatedAt           time.Time `json:"created_at"`
}

func toProductCardResponse(p domain.ProductCard) productCardResponse {
	return productCardResponse{ID: p.ID, Name: p.Name, Category: p.Category}
}

func toRequestViewResponse(r domain.RequestView) requestViewResponse {
	return requestViewResponse{
		ID:                  r.ID,
		ProductID:           r.ProductID,
		DesiredInstallments: r.DesiredInstallments,
		DesiredDownPayment:  r.DesiredDownPayment.String(),
		Note:                r.Note,
		Status:              r.Status,
		ContractID:          r.ContractID,
		CreatedAt:           r.CreatedAt,
	}
}

func toContractDetailResponse(d domain.ContractDetail) contractDetailResponse {
	schedule := make([]installmentLineResponse, 0, len(d.Installments))
	for _, l := range d.Installments {
		schedule = append(schedule, installmentLineResponse{
			Number:  l.Number,
			DueDate: l.DueDate.Format(dateLayout),
			Amount:  l.Amount.String(),
			Status:  l.Status,
		})
	}
	payments := make([]paymentLineResponse, 0, len(d.Payments))
	for _, p := range d.Payments {
		payments = append(payments, paymentLineResponse{Amount: p.Amount.String(), PaidAt: p.PaidAt})
	}
	resp := contractDetailResponse{
		ID:             d.ID,
		ProductID:      d.ProductID,
		SalePrice:      d.SalePrice.String(),
		DownPayment:    d.DownPayment.String(),
		FinancedAmount: d.FinancedAmount.String(),
		Outstanding:    d.Outstanding.String(),
		PaidAmount:     d.PaidAmount.String(),
		Status:         d.Status,
		Cadence:        d.Cadence,
		StartDate:      d.StartDate.Format(dateLayout),
		HasOverdue:     d.HasOverdue,
		HasNext:        d.HasNext,
		Schedule:       schedule,
		Payments:       payments,
		CreatedAt:      d.CreatedAt,
	}
	if d.HasNext {
		resp.NextDueDate = d.NextDueDate.Format(dateLayout)
		resp.NextDueAmount = d.NextDueAmount.String()
	}
	return resp
}
