package http

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
)

type createExpenseRequest struct {
	Category string `json:"category" validate:"required"`
	Amount   string `json:"amount" validate:"required"` // decimal string, e.g. "5000.00"
	Note     string `json:"note"`
	SpentAt  string `json:"spent_at"` // YYYY-MM-DD, optional (defaults to today)
}

type expenseResponse struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Amount    string    `json:"amount"` // decimal string
	Note      string    `json:"note"`
	SpentAt   string    `json:"spent_at"` // YYYY-MM-DD
	CreatedAt time.Time `json:"created_at"`
}

type saleLineResponse struct {
	ContractID string    `json:"contract_id"`
	ProductID  string    `json:"product_id"`
	SalePrice  string    `json:"sale_price"`
	CostPrice  string    `json:"cost_price"`
	Profit     string    `json:"profit"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type reportResponse struct {
	Revenue       string             `json:"revenue"`
	CostOfGoods   string             `json:"cost_of_goods"`
	GrossProfit   string             `json:"gross_profit"`
	OtherExpenses string             `json:"other_expenses"`
	NetProfit     string             `json:"net_profit"`
	SalesCount    int                `json:"sales_count"`
	ExpensesCount int                `json:"expenses_count"`
	Sales         []saleLineResponse `json:"sales"`
}

const dateLayout = "2006-01-02"

func toExpenseResponse(e domain.Expense) expenseResponse {
	return expenseResponse{
		ID:        e.ID(),
		Category:  e.Category(),
		Amount:    e.Amount().String(),
		Note:      e.Note(),
		SpentAt:   e.SpentAt().Format(dateLayout),
		CreatedAt: e.CreatedAt(),
	}
}

func toReportResponse(r domain.Report, sales []domain.Sale) reportResponse {
	lines := make([]saleLineResponse, 0, len(sales))
	for _, s := range sales {
		profit, _ := s.Profit() // single-currency: no mismatch
		lines = append(lines, saleLineResponse{
			ContractID: s.ContractID,
			ProductID:  s.ProductID,
			SalePrice:  s.SalePrice.String(),
			CostPrice:  s.CostPrice.String(),
			Profit:     profit.String(),
			Status:     s.Status,
			CreatedAt:  s.CreatedAt,
		})
	}
	return reportResponse{
		Revenue:       r.Revenue.String(),
		CostOfGoods:   r.CostOfGoods.String(),
		GrossProfit:   r.GrossProfit.String(),
		OtherExpenses: r.OtherExpenses.String(),
		NetProfit:     r.NetProfit.String(),
		SalesCount:    r.SalesCount,
		ExpensesCount: r.ExpensesCount,
		Sales:         lines,
	}
}
