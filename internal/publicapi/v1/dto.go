package v1

import (
	"time"

	"github.com/shopspring/decimal"

	financingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
)

// CreateContractRequest is the body of POST /api/v1/contracts. It references an
// existing client and product; money fields are decimal strings. Markup is given
// as an amount (the fixed murabaha surcharge).
type CreateContractRequest struct {
	ClientID     string `json:"client_id" validate:"required,uuid" example:"6f9619ff-8b86-d011-b42d-00cf4fc964ff"`
	ProductID    string `json:"product_id" validate:"required,uuid" example:"7a1b2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d"`
	CostPrice    string `json:"cost_price" validate:"required" example:"100000.00"`
	Markup       string `json:"markup" validate:"required" example:"20000.00"`
	DownPayment  string `json:"down_payment" example:"30000.00"`
	Installments int    `json:"installments" validate:"required,min=1" example:"6"`
	Cadence      string `json:"cadence" validate:"required,oneof=weekly monthly" example:"monthly"`
	StartDate    string `json:"start_date" validate:"required" example:"2026-07-01"`
}

// ContractResponse is the compact public representation of a created contract.
type ContractResponse struct {
	ID             string `json:"id"`
	Status         string `json:"status" example:"active"`
	ClientID       string `json:"client_id"`
	ProductID      string `json:"product_id"`
	CostPrice      string `json:"cost_price" example:"100000.00"`
	Markup         string `json:"markup" example:"20000.00"`
	SalePrice      string `json:"sale_price" example:"120000.00"`
	DownPayment    string `json:"down_payment" example:"30000.00"`
	FinancedAmount string `json:"financed_amount" example:"90000.00"`
	Outstanding    string `json:"outstanding" example:"90000.00"`
	Installments   int    `json:"installments" example:"6"`
	Cadence        string `json:"cadence" example:"monthly"`
	StartDate      string `json:"start_date" example:"2026-07-01"`
}

// InstallmentDTO is one planned payment with its derived status.
type InstallmentDTO struct {
	Number  int    `json:"number" example:"1"`
	DueDate string `json:"due_date" example:"2026-07-01"`
	Amount  string `json:"amount" example:"15000.00"`
	Status  string `json:"status" example:"pending"`
}

// PaymentDTO is one registered payment.
type PaymentDTO struct {
	Amount string    `json:"amount" example:"15000.00"`
	PaidAt time.Time `json:"paid_at"`
}

// PaymentsResponse is the body of GET /api/v1/contracts/{id}/payments.
type PaymentsResponse struct {
	ContractID      string           `json:"contract_id"`
	Status          string           `json:"status" example:"active"`
	SalePrice       string           `json:"sale_price" example:"120000.00"`
	Outstanding     string           `json:"outstanding" example:"75000.00"`
	PaidAmount      string           `json:"paid_amount" example:"15000.00"`
	ProgressPercent string           `json:"progress_percent" example:"16.67"`
	HasOverdue      bool             `json:"has_overdue" example:"false"`
	Schedule        []InstallmentDTO `json:"schedule"`
	Payments        []PaymentDTO     `json:"payments"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody carries a machine-readable code and a message.
type ErrorBody struct {
	Code    string `json:"code" example:"invalid_input"`
	Message string `json:"message" example:"down payment must be less than sale price"`
}

const dateLayout = "2006-01-02"

func toContractResponse(c *financingdomain.Contract) ContractResponse {
	return ContractResponse{
		ID:             c.ID(),
		Status:         c.Status().String(),
		ClientID:       c.ClientID(),
		ProductID:      c.ProductID(),
		CostPrice:      c.CostPrice().String(),
		Markup:         c.Markup().Money().String(),
		SalePrice:      c.SalePrice().String(),
		DownPayment:    c.DownPayment().String(),
		FinancedAmount: c.FinancedAmount().String(),
		Outstanding:    c.Outstanding().String(),
		Installments:   c.InstallmentsCount(),
		Cadence:        c.Cadence().String(),
		StartDate:      c.StartDate().Format(dateLayout),
	}
}

func toPaymentsResponse(c *financingdomain.Contract, asOf time.Time) PaymentsResponse {
	views := c.Installments(asOf)
	schedule := make([]InstallmentDTO, 0, len(views))
	for _, v := range views {
		schedule = append(schedule, InstallmentDTO{
			Number:  v.Number,
			DueDate: v.DueDate.Format(dateLayout),
			Amount:  v.Amount.String(),
			Status:  v.Status.String(),
		})
	}
	payments := make([]PaymentDTO, 0, len(c.Payments()))
	for _, p := range c.Payments() {
		payments = append(payments, PaymentDTO{Amount: p.Amount().String(), PaidAt: p.PaidAt()})
	}

	progress := "0.00"
	if !c.FinancedAmount().Amount().IsZero() {
		progress = c.PaidAmount().Amount().Mul(decimal.NewFromInt(100)).Div(c.FinancedAmount().Amount()).Round(2).String()
	}

	return PaymentsResponse{
		ContractID:      c.ID(),
		Status:          c.Status().String(),
		SalePrice:       c.SalePrice().String(),
		Outstanding:     c.Outstanding().String(),
		PaidAmount:      c.PaidAmount().String(),
		ProgressPercent: progress,
		HasOverdue:      c.HasOverdue(asOf),
		Schedule:        schedule,
		Payments:        payments,
	}
}
