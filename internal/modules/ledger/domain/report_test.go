package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

func rub(t *testing.T, s string) money.Money {
	t.Helper()
	m, err := money.FromString(s, "RUB")
	require.NoError(t, err)
	return m
}

func sale(t *testing.T, salePrice, costPrice, status string) Sale {
	return Sale{SalePrice: rub(t, salePrice), CostPrice: rub(t, costPrice), Status: status}
}

func expense(t *testing.T, amount string) Expense {
	e, err := NewExpense("e1", "o1", "ремонт", rub(t, amount), "", time.Now())
	require.NoError(t, err)
	return e
}

func TestBuildReport(t *testing.T) {
	t.Run("revenue minus cost gives gross profit; minus expenses gives net", func(t *testing.T) {
		sales := []Sale{
			sale(t, "120000.00", "85000.00", SaleStatusActive),    // profit 35000
			sale(t, "65000.00", "50000.00", SaleStatusCompleted),  // profit 15000
		}
		expenses := []Expense{expense(t, "10000.00"), expense(t, "5000.00")} // 15000

		r, err := BuildReport(sales, expenses, "RUB")
		require.NoError(t, err)
		assert.Equal(t, "185000.00", r.Revenue.String())
		assert.Equal(t, "135000.00", r.CostOfGoods.String())
		assert.Equal(t, "50000.00", r.GrossProfit.String())   // 185000 − 135000
		assert.Equal(t, "15000.00", r.OtherExpenses.String())
		assert.Equal(t, "35000.00", r.NetProfit.String())     // 50000 − 15000
		assert.Equal(t, 2, r.SalesCount)
		assert.Equal(t, 2, r.ExpensesCount)
	})

	t.Run("cancelled sales are excluded from revenue and cost", func(t *testing.T) {
		sales := []Sale{
			sale(t, "120000.00", "85000.00", SaleStatusActive),
			sale(t, "999999.00", "999999.00", SaleStatusCancelled), // ignored entirely
		}
		r, err := BuildReport(sales, nil, "RUB")
		require.NoError(t, err)
		assert.Equal(t, "120000.00", r.Revenue.String())
		assert.Equal(t, "85000.00", r.CostOfGoods.String())
		assert.Equal(t, "35000.00", r.GrossProfit.String())
		assert.Equal(t, 1, r.SalesCount)
	})

	t.Run("expenses can exceed gross profit producing a negative net", func(t *testing.T) {
		sales := []Sale{sale(t, "100000.00", "90000.00", SaleStatusActive)} // gross 10000
		expenses := []Expense{expense(t, "25000.00")}
		r, err := BuildReport(sales, expenses, "RUB")
		require.NoError(t, err)
		assert.Equal(t, "10000.00", r.GrossProfit.String())
		assert.Equal(t, "-15000.00", r.NetProfit.String())
		assert.True(t, r.NetProfit.IsNegative())
	})

	t.Run("no data yields zeros", func(t *testing.T) {
		r, err := BuildReport(nil, nil, "RUB")
		require.NoError(t, err)
		assert.Equal(t, "0.00", r.Revenue.String())
		assert.Equal(t, "0.00", r.NetProfit.String())
		assert.Equal(t, 0, r.SalesCount)
	})
}

func TestSaleProfit(t *testing.T) {
	p, err := sale(t, "120000.00", "85000.00", SaleStatusActive).Profit()
	require.NoError(t, err)
	assert.Equal(t, "35000.00", p.String())
}

func TestNewExpense(t *testing.T) {
	t.Run("valid trims category and note", func(t *testing.T) {
		e, err := NewExpense("e1", "o1", "  Ремонт витрины  ", rub(t, "5000.00"), "  стекло  ", time.Time{})
		require.NoError(t, err)
		assert.Equal(t, "Ремонт витрины", e.Category())
		assert.Equal(t, "стекло", e.Note())
		assert.False(t, e.SpentAt().IsZero()) // zero spentAt defaults to now
	})

	cases := []struct {
		name     string
		id       string
		orgID    string
		category string
		amount   string
		wantErr  error
	}{
		{"empty id", "", "o1", "ремонт", "100.00", ErrExpenseIDRequired},
		{"empty org", "e1", "", "ремонт", "100.00", ErrOrgIDRequired},
		{"empty category", "e1", "o1", "   ", "100.00", ErrCategoryRequired},
		{"zero amount", "e1", "o1", "ремонт", "0.00", ErrAmountNotPositive},
		{"negative amount", "e1", "o1", "ремонт", "-50.00", ErrAmountNotPositive},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewExpense(tc.id, tc.orgID, tc.category, rub(t, tc.amount), "", time.Now())
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
