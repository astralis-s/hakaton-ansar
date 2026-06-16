package domain

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Sale statuses the ledger cares about (mirrors financing contract status, kept
// as plain strings so the ledger domain does not import financing).
const (
	SaleStatusActive    = "active"
	SaleStatusCompleted = "completed"
	SaleStatusCancelled = "cancelled"
)

// Sale is the ledger's read model of a financing contract: what a good was sold
// for (SalePrice) versus what it was bought for (CostPrice). Income on a sale =
// SalePrice − CostPrice (the murabaha markup). Cancelled sales do not count.
type Sale struct {
	ContractID string
	ProductID  string
	SalePrice  money.Money
	CostPrice  money.Money
	Status     string
	CreatedAt  time.Time
}

// Counts reports whether this sale contributes to the P&L (i.e. not cancelled).
func (s Sale) Counts() bool { return s.Status != SaleStatusCancelled }

// Profit is the income on a single sale: SalePrice − CostPrice.
func (s Sale) Profit() (money.Money, error) { return s.SalePrice.Sub(s.CostPrice) }

// Report is the income/expense summary (P&L) for an organization:
//
//	Revenue       = Σ sale price of non-cancelled sales   (за сколько продано)
//	CostOfGoods   = Σ cost price of non-cancelled sales   (за сколько куплено)
//	GrossProfit   = Revenue − CostOfGoods                 (доход = продажа − покупка)
//	OtherExpenses = Σ manual expenses (rent, repairs, …)
//	NetProfit     = GrossProfit − OtherExpenses           (чистая прибыль)
type Report struct {
	Revenue       money.Money
	CostOfGoods   money.Money
	GrossProfit   money.Money
	OtherExpenses money.Money
	NetProfit     money.Money
	SalesCount    int
	ExpensesCount int
}

// BuildReport aggregates sales and manual expenses into a P&L. All amounts are
// in the given currency (single-currency MVP); a currency mismatch is an error.
func BuildReport(sales []Sale, expenses []Expense, currency string) (Report, error) {
	revenue := money.Zero(currency)
	cost := money.Zero(currency)
	salesCount := 0
	for _, s := range sales {
		if !s.Counts() {
			continue
		}
		var err error
		if revenue, err = revenue.Add(s.SalePrice); err != nil {
			return Report{}, err
		}
		if cost, err = cost.Add(s.CostPrice); err != nil {
			return Report{}, err
		}
		salesCount++
	}

	other := money.Zero(currency)
	for _, e := range expenses {
		var err error
		if other, err = other.Add(e.Amount()); err != nil {
			return Report{}, err
		}
	}

	gross, err := revenue.Sub(cost)
	if err != nil {
		return Report{}, err
	}
	net, err := gross.Sub(other)
	if err != nil {
		return Report{}, err
	}

	return Report{
		Revenue:       revenue,
		CostOfGoods:   cost,
		GrossProfit:   gross,
		OtherExpenses: other,
		NetProfit:     net,
		SalesCount:    salesCount,
		ExpensesCount: len(expenses),
	}, nil
}
