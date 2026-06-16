package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// GetReport builds the income/expense P&L for an organization from its sales
// (financing contracts) and its manual expenses. It returns the summary plus the
// per-sale breakdown so the UI can show where the income came from.
type GetReport struct {
	sales    domain.SalesReader
	expenses domain.ExpenseRepository
}

func NewGetReport(sales domain.SalesReader, expenses domain.ExpenseRepository) *GetReport {
	return &GetReport{sales: sales, expenses: expenses}
}

func (uc *GetReport) Execute(ctx context.Context, orgID string) (domain.Report, []domain.Sale, error) {
	sales, err := uc.sales.ListSales(ctx, orgID)
	if err != nil {
		return domain.Report{}, nil, err
	}
	expenses, err := uc.expenses.ListByOrg(ctx, orgID)
	if err != nil {
		return domain.Report{}, nil, err
	}
	report, err := domain.BuildReport(sales, expenses, money.DefaultCurrency)
	if err != nil {
		return domain.Report{}, nil, err
	}
	return report, sales, nil
}
