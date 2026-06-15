package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func TestBuildDashboard(t *testing.T) {
	// A: active, started long ago, unpaid → overdue. Outstanding 90000.
	a, err := NewContract(NewContractParams{ID: "A", OrgID: "o1", ClientID: "clA", ProductID: "p1",
		CostPrice: rub(t, "100000.00"), Markup: mkMarkup(t, "20000.00"), DownPayment: rub(t, "30000.00"),
		Installments: 6, Cadence: CadenceMonthly, StartDate: date(2026, 1, 1)})
	require.NoError(t, err)
	require.NoError(t, a.Activate())

	// B: active, first installment due this week (start Thu 2026-06-18), one payment this week.
	b, err := NewContract(NewContractParams{ID: "B", OrgID: "o1", ClientID: "clB", ProductID: "p2",
		CostPrice: rub(t, "40000.00"), Markup: mkMarkup(t, "0.00"), DownPayment: rub(t, "0.00"),
		Installments: 4, Cadence: CadenceWeekly, StartDate: date(2026, 6, 18)})
	require.NoError(t, err)
	require.NoError(t, b.Activate())
	require.NoError(t, b.RegisterPayment("pay", rub(t, "5000.00"), date(2026, 6, 16)))

	// C: cancelled → excluded from everything.
	c, err := NewContract(NewContractParams{ID: "C", OrgID: "o1", ClientID: "clA", ProductID: "p3",
		CostPrice: rub(t, "10000.00"), Markup: mkMarkup(t, "0.00"), DownPayment: rub(t, "0.00"),
		Installments: 2, Cadence: CadenceMonthly, StartDate: date(2026, 1, 1)})
	require.NoError(t, err)
	require.NoError(t, c.Activate())
	require.NoError(t, c.Cancel())

	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC) // Wednesday
	names := map[string]string{"clA": "Иван", "clB": "Хеда"}

	res, err := BuildDashboard(now, []*Contract{a, b, c}, names, "RUB")
	require.NoError(t, err)

	assert.Equal(t, 2, res.ActiveContracts, "C is cancelled, excluded")
	assert.Equal(t, "125000.00", res.PortfolioOutstanding.String(), "A 90000 + B 35000")
	assert.Equal(t, "10000.00", res.WeekExpected.String(), "B installment due 06-18")
	assert.Equal(t, "5000.00", res.WeekCollected.String(), "B payment on 06-16")
	assert.Equal(t, "50", res.CollectionRatePercent.String(), "5000/10000 = 50%")

	require.Len(t, res.Overdue, 1)
	assert.Equal(t, "A", res.Overdue[0].ContractID)
	assert.Equal(t, "Иван", res.Overdue[0].ClientName)
	assert.Equal(t, "90000.00", res.Overdue[0].Outstanding.String())
	assert.Equal(t, 167, res.Overdue[0].DaysOverdue, "2026-01-01 → 2026-06-17")

	require.Len(t, res.Upcoming, 1)
	assert.Equal(t, "B", res.Upcoming[0].ContractID)
	assert.Equal(t, "Хеда", res.Upcoming[0].ClientName)
	assert.Equal(t, "10000.00", res.Upcoming[0].Amount.String())
	assert.Equal(t, InstallmentPartiallyPaid, res.Upcoming[0].Status)
}

func TestBuildDashboard_EmptyOrg(t *testing.T) {
	res, err := BuildDashboard(time.Now(), nil, map[string]string{}, "RUB")
	require.NoError(t, err)
	assert.Equal(t, 0, res.ActiveContracts)
	assert.Equal(t, "0.00", res.PortfolioOutstanding.String())
	assert.Equal(t, "0", res.CollectionRatePercent.String(), "no expected → rate 0")
	assert.Empty(t, res.Overdue)
	assert.Empty(t, res.Upcoming)
}
