package app

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/document"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

const docDateTime = "02.01.2006, 15:04"
const docDate = "02.01.2006"

// BuildReportDoc gathers the denormalized data for the one-page finance report.
type BuildReportDoc struct {
	sales    domain.SalesReader
	expenses domain.ExpenseRepository
	orgs     domain.OrgReader
}

func NewBuildReportDoc(sales domain.SalesReader, expenses domain.ExpenseRepository, orgs domain.OrgReader) *BuildReportDoc {
	return &BuildReportDoc{sales: sales, expenses: expenses, orgs: orgs}
}

func (uc *BuildReportDoc) Execute(ctx context.Context, orgID string) (document.FinanceReport, error) {
	sales, err := uc.sales.ListSales(ctx, orgID)
	if err != nil {
		return document.FinanceReport{}, err
	}
	expenses, err := uc.expenses.ListByOrg(ctx, orgID)
	if err != nil {
		return document.FinanceReport{}, err
	}
	report, err := domain.BuildReport(sales, expenses, money.DefaultCurrency)
	if err != nil {
		return document.FinanceReport{}, err
	}

	doc := document.FinanceReport{
		GeneratedAt:   time.Now().Format(docDateTime),
		Revenue:       report.Revenue.String(),
		CostOfGoods:   report.CostOfGoods.String(),
		GrossProfit:   report.GrossProfit.String(),
		OtherExpenses: report.OtherExpenses.String(),
		NetProfit:     report.NetProfit.String(),
		NetNegative:   report.NetProfit.IsNegative(),
		SalesCount:    report.SalesCount,
		ExpensesCount: report.ExpensesCount,
		AvgSale:       avgSale(report.Revenue, report.SalesCount).String(),
	}
	if name, err := uc.orgs.Name(ctx, orgID); err == nil {
		doc.OrgName = name
	}
	for _, e := range expenses {
		doc.Expenses = append(doc.Expenses, document.ReportExpense{
			Category: e.Category(),
			Amount:   e.Amount().String(),
			Date:     e.SpentAt().Format(docDate),
		})
	}
	return doc, nil
}

// avgSale is revenue / salesCount, rounded to 2 places (zero when no sales).
func avgSale(revenue money.Money, count int) money.Money {
	if count <= 0 {
		return money.Zero(money.DefaultCurrency)
	}
	avg := revenue.Amount().DivRound(decimal.NewFromInt(int64(count)), 2)
	return money.New(avg, money.DefaultCurrency)
}
