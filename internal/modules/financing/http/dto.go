package http

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
)

const dateLayout = "2006-01-02"

// --- Requests ---------------------------------------------------------------

// previewRequest / createContractRequest share the same terms. Markup is given
// either as an amount or as a percent of the cost price (exactly one).
type previewRequest struct {
	CostPrice     string `json:"cost_price" validate:"required"`
	MarkupAmount  string `json:"markup_amount"`
	MarkupPercent string `json:"markup_percent"`
	DownPayment   string `json:"down_payment"`
	Installments  int    `json:"installments" validate:"required,min=1"`
	Cadence       string `json:"cadence" validate:"required,oneof=weekly monthly"`
	StartDate     string `json:"start_date" validate:"required"`
}

type createContractRequest struct {
	ClientID      string `json:"client_id" validate:"required,uuid"`
	ProductID     string `json:"product_id" validate:"required,uuid"`
	CostPrice     string `json:"cost_price" validate:"required"`
	MarkupAmount  string `json:"markup_amount"`
	MarkupPercent string `json:"markup_percent"`
	DownPayment   string `json:"down_payment"`
	Installments  int    `json:"installments" validate:"required,min=1"`
	Cadence       string `json:"cadence" validate:"required,oneof=weekly monthly"`
	StartDate     string `json:"start_date" validate:"required"`
}

type registerPaymentRequest struct {
	Amount string `json:"amount" validate:"required"`
}

// --- Responses --------------------------------------------------------------

type installmentDTO struct {
	Number  int    `json:"number"`
	DueDate string `json:"due_date"`
	Amount  string `json:"amount"`
}

type installmentViewDTO struct {
	Number  int    `json:"number"`
	DueDate string `json:"due_date"`
	Amount  string `json:"amount"`
	Status  string `json:"status"`
}

type paymentDTO struct {
	ID     string    `json:"id"`
	Amount string    `json:"amount"`
	PaidAt time.Time `json:"paid_at"`
}

type comparisonDTO struct {
	MurabahaTotal     string `json:"murabaha_total"`
	ConventionalTotal string `json:"conventional_total"`
	Overpayment       string `json:"overpayment"`
	AnnualRatePercent string `json:"annual_rate_percent"`
}

type previewResponse struct {
	SalePrice      string           `json:"sale_price"`
	FinancedAmount string           `json:"financed_amount"`
	Schedule       []installmentDTO `json:"schedule"`
	Comparison     comparisonDTO    `json:"comparison"`
}

type contractResponse struct {
	ID              string               `json:"id"`
	OrgID           string               `json:"org_id"`
	ClientID        string               `json:"client_id"`
	ProductID       string               `json:"product_id"`
	CostPrice       string               `json:"cost_price"`
	Markup          string               `json:"markup"`
	SalePrice       string               `json:"sale_price"`
	DownPayment     string               `json:"down_payment"`
	FinancedAmount  string               `json:"financed_amount"`
	Outstanding     string               `json:"outstanding"`
	PaidAmount      string               `json:"paid_amount"`
	ProgressPercent string               `json:"progress_percent"`
	Status          string               `json:"status"`
	Cadence         string               `json:"cadence"`
	StartDate       string               `json:"start_date"`
	HasOverdue      bool                 `json:"has_overdue"`
	Schedule        []installmentViewDTO `json:"schedule"`
	Payments        []paymentDTO         `json:"payments"`
	CreatedAt       time.Time            `json:"created_at"`
}

type contractSummaryDTO struct {
	ID              string    `json:"id"`
	ClientID        string    `json:"client_id"`
	ProductID       string    `json:"product_id"`
	SalePrice       string    `json:"sale_price"`
	FinancedAmount  string    `json:"financed_amount"`
	Outstanding     string    `json:"outstanding"`
	ProgressPercent string    `json:"progress_percent"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// --- Mappers ----------------------------------------------------------------

func progressPercent(c *domain.Contract) string {
	financed := c.FinancedAmount().Amount()
	if financed.IsZero() {
		return "0.00"
	}
	return c.PaidAmount().Amount().Mul(decimal.NewFromInt(100)).Div(financed).Round(2).String()
}

func toPreviewResponse(p domain.PreviewResult) previewResponse {
	schedule := make([]installmentDTO, 0, len(p.Schedule))
	for _, inst := range p.Schedule {
		schedule = append(schedule, installmentDTO{
			Number:  inst.Number(),
			DueDate: inst.DueDate().Format(dateLayout),
			Amount:  inst.Amount().String(),
		})
	}
	return previewResponse{
		SalePrice:      p.SalePrice.String(),
		FinancedAmount: p.FinancedAmount.String(),
		Schedule:       schedule,
		Comparison: comparisonDTO{
			MurabahaTotal:     p.Comparison.MurabahaTotal.String(),
			ConventionalTotal: p.Comparison.ConventionalTotal.String(),
			Overpayment:       p.Comparison.Overpayment.String(),
			AnnualRatePercent: p.Comparison.AnnualRatePercent.String(),
		},
	}
}

func toContractResponse(c *domain.Contract, asOf time.Time) contractResponse {
	views := c.Installments(asOf)
	schedule := make([]installmentViewDTO, 0, len(views))
	for _, v := range views {
		schedule = append(schedule, installmentViewDTO{
			Number:  v.Number,
			DueDate: v.DueDate.Format(dateLayout),
			Amount:  v.Amount.String(),
			Status:  v.Status.String(),
		})
	}
	payments := make([]paymentDTO, 0, len(c.Payments()))
	for _, p := range c.Payments() {
		payments = append(payments, paymentDTO{ID: p.ID(), Amount: p.Amount().String(), PaidAt: p.PaidAt()})
	}
	return contractResponse{
		ID:              c.ID(),
		OrgID:           c.OrgID(),
		ClientID:        c.ClientID(),
		ProductID:       c.ProductID(),
		CostPrice:       c.CostPrice().String(),
		Markup:          c.Markup().Money().String(),
		SalePrice:       c.SalePrice().String(),
		DownPayment:     c.DownPayment().String(),
		FinancedAmount:  c.FinancedAmount().String(),
		Outstanding:     c.Outstanding().String(),
		PaidAmount:      c.PaidAmount().String(),
		ProgressPercent: progressPercent(c),
		Status:          c.Status().String(),
		Cadence:         c.Cadence().String(),
		StartDate:       c.StartDate().Format(dateLayout),
		HasOverdue:      c.HasOverdue(asOf),
		Schedule:        schedule,
		Payments:        payments,
		CreatedAt:       c.CreatedAt(),
	}
}

func toContractSummaryDTO(s domain.ContractSummary) contractSummaryDTO {
	progress := "0.00"
	if !s.FinancedAmount.Amount().IsZero() {
		paid, _ := s.FinancedAmount.Sub(s.Outstanding)
		progress = paid.Amount().Mul(decimal.NewFromInt(100)).Div(s.FinancedAmount.Amount()).Round(2).String()
	}
	return contractSummaryDTO{
		ID:              s.ID,
		ClientID:        s.ClientID,
		ProductID:       s.ProductID,
		SalePrice:       s.SalePrice.String(),
		FinancedAmount:  s.FinancedAmount.String(),
		Outstanding:     s.Outstanding.String(),
		ProgressPercent: progress,
		Status:          s.Status.String(),
		CreatedAt:       s.CreatedAt,
	}
}
