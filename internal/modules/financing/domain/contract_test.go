package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
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

func mkMarkup(t *testing.T, s string) Markup {
	t.Helper()
	mk, err := NewMarkup(rub(t, s))
	require.NoError(t, err)
	return mk
}

// future start so nothing is overdue unless a test wants it.
func futureStart() time.Time { return time.Now().AddDate(1, 0, 0) }

func baseParams(t *testing.T) NewContractParams {
	t.Helper()
	return NewContractParams{
		ID: "c1", OrgID: "o1", ClientID: "cl1", ProductID: "p1",
		CostPrice:    rub(t, "100000.00"),
		Markup:       mkMarkup(t, "20000.00"),
		DownPayment:  rub(t, "30000.00"),
		Installments: 6,
		Cadence:      CadenceMonthly,
		StartDate:    futureStart(),
	}
}

func sumScheduleCents(s []Installment) int64 {
	var total int64
	for _, i := range s {
		total += i.Amount().Cents()
	}
	return total
}

// 1. Even division.
func TestNewContract_EvenDivision(t *testing.T) {
	c, err := NewContract(baseParams(t))
	require.NoError(t, err)

	assert.Equal(t, "120000.00", c.SalePrice().String())
	assert.Equal(t, "90000.00", c.FinancedAmount().String())
	assert.Equal(t, "90000.00", c.Outstanding().String())
	require.Len(t, c.Schedule(), 6)
	for _, inst := range c.Schedule() {
		assert.Equal(t, "15000.00", inst.Amount().String())
	}
	// kopecks reconcile: Σ == financed, Down + Σ == sale.
	assert.Equal(t, c.FinancedAmount().Cents(), sumScheduleCents(c.Schedule()))
	assert.Equal(t, int64(12000000), c.DownPayment().Cents()+sumScheduleCents(c.Schedule()))
}

// 2. Division with remainder — extra kopeck onto the earliest payment(s).
func TestNewContract_DivisionWithRemainder(t *testing.T) {
	p := baseParams(t)
	p.CostPrice = rub(t, "100000.00")
	p.Markup = mkMarkup(t, "0.00")
	p.DownPayment = rub(t, "0.00")
	p.Installments = 3

	c, err := NewContract(p)
	require.NoError(t, err)

	got := []string{}
	for _, inst := range c.Schedule() {
		got = append(got, inst.Amount().String())
	}
	assert.Equal(t, []string{"33333.34", "33333.33", "33333.33"}, got)
	assert.Equal(t, int64(10000000), sumScheduleCents(c.Schedule()))
	assert.Equal(t, c.FinancedAmount().Cents(), sumScheduleCents(c.Schedule()))
}

// 3. Zero down payment → financed == sale.
func TestNewContract_ZeroDownPayment(t *testing.T) {
	p := baseParams(t)
	p.DownPayment = rub(t, "0.00")
	c, err := NewContract(p)
	require.NoError(t, err)

	assert.Equal(t, c.SalePrice().String(), c.FinancedAmount().String())
	assert.Equal(t, "120000.00", c.FinancedAmount().String())
	assert.Equal(t, c.FinancedAmount().Cents(), sumScheduleCents(c.Schedule()))
}

// 4. Markup from percent is fixed as an amount and never recomputed.
func TestNewContract_MarkupFromPercent(t *testing.T) {
	cost := rub(t, "100000.00")
	mk, err := NewMarkupFromPercent(cost, decimal.NewFromInt(10))
	require.NoError(t, err)
	assert.Equal(t, "10000.00", mk.Money().String())

	p := baseParams(t)
	p.CostPrice = cost
	p.Markup = mk
	p.DownPayment = rub(t, "0.00")
	c, err := NewContract(p)
	require.NoError(t, err)

	assert.Equal(t, "110000.00", c.SalePrice().String())
	// markup stays fixed regardless of anything else
	assert.Equal(t, "10000.00", c.Markup().Money().String())
}

func activeContract(t *testing.T, p NewContractParams) *Contract {
	t.Helper()
	c, err := NewContract(p)
	require.NoError(t, err)
	require.NoError(t, c.Activate())
	return c
}

// 5. Full repayment by installments → Outstanding 0, Completed.
func TestRegisterPayment_FullRepayment(t *testing.T) {
	c := activeContract(t, baseParams(t)) // financed 90000, 6×15000
	for i := 0; i < 6; i++ {
		require.NoError(t, c.RegisterPayment("pay", rub(t, "15000.00"), time.Now()))
	}
	assert.True(t, c.Outstanding().IsZero())
	assert.Equal(t, StatusCompleted, c.Status())

	// further payment rejected (not active)
	err := c.RegisterPayment("pay", rub(t, "1.00"), time.Now())
	require.ErrorIs(t, err, ErrContractNotActive)
}

// 6. Partial payment not aligned to an installment.
func TestRegisterPayment_PartialStatuses(t *testing.T) {
	c := activeContract(t, baseParams(t)) // 6×15000, financed 90000
	require.NoError(t, c.RegisterPayment("pay", rub(t, "20000.00"), time.Now()))

	assert.Equal(t, "70000.00", c.Outstanding().String())

	asOf := time.Now() // before all (future) due dates
	views := c.Installments(asOf)
	var paid, partial, pending int
	for _, v := range views {
		switch v.Status {
		case InstallmentPaid:
			paid++
		case InstallmentPartiallyPaid:
			partial++
		case InstallmentPending:
			pending++
		}
	}
	assert.Equal(t, 1, paid, "installment 1 fully covered")
	assert.Equal(t, 1, partial, "exactly one partially paid")
	assert.Equal(t, 4, pending, "the rest pending")
}

// 7. Overdue does not change the debt.
func TestOverdue_DoesNotChangeDebt(t *testing.T) {
	p := baseParams(t)
	p.StartDate = time.Now().AddDate(-1, 0, 0) // a year ago → all due dates passed
	c := activeContract(t, p)

	saleBefore := c.SalePrice().String()
	outstandingBefore := c.Outstanding().String()

	assert.True(t, c.HasOverdue(time.Now()))
	assert.Equal(t, saleBefore, c.SalePrice().String())
	assert.Equal(t, outstandingBefore, c.Outstanding().String())
	assert.Equal(t, "90000.00", c.Outstanding().String())
}

// 8. Overdue never changes the debt (anti-riba; covered by TestOverdue_DoesNotChangeDebt).

// 9. Early settlement clears the balance with no penalty.
func TestSettleEarly_NoPenalty(t *testing.T) {
	c := activeContract(t, baseParams(t))
	require.NoError(t, c.RegisterPayment("p", rub(t, "15000.00"), time.Now())) // outstanding 75000

	saleBefore := c.SalePrice().String()
	require.NoError(t, c.SettleEarly("settle", time.Now()))

	assert.True(t, c.Outstanding().IsZero())
	assert.Equal(t, StatusCompleted, c.Status())
	assert.Equal(t, saleBefore, c.SalePrice().String(), "sale price unchanged (no penalty)")
}

// 10. Preview equals creation — same schedule, no persistence.
func TestPreview_EqualsCreation(t *testing.T) {
	p := baseParams(t)
	in := PreviewInput{
		CostPrice:    p.CostPrice,
		Markup:       p.Markup,
		DownPayment:  p.DownPayment,
		Installments: p.Installments,
		Cadence:      p.Cadence,
		StartDate:    p.StartDate,
	}
	preview, err := Preview(in, decimal.NewFromInt(28))
	require.NoError(t, err)

	c, err := NewContract(p)
	require.NoError(t, err)

	assert.Equal(t, c.SalePrice().String(), preview.SalePrice.String())
	assert.Equal(t, c.FinancedAmount().String(), preview.FinancedAmount.String())
	require.Equal(t, len(c.Schedule()), len(preview.Schedule))
	for i := range c.Schedule() {
		assert.Equal(t, c.Schedule()[i].Amount().String(), preview.Schedule[i].Amount().String())
		assert.True(t, c.Schedule()[i].DueDate().Equal(preview.Schedule[i].DueDate()))
	}
	// comparison shows riba you avoid
	assert.True(t, preview.Comparison.Overpayment.IsPositive())
	assert.Equal(t, c.SalePrice().String(), preview.Comparison.MurabahaTotal.String())
}

// 11. Invalid inputs are rejected.
func TestNewContract_InvalidInputs(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(p *NewContractParams)
		wantErr error
	}{
		{"down >= sale", func(p *NewContractParams) { p.DownPayment = rub(t, "120000.00") }, ErrDownPaymentTooLarge},
		{"installments < 1", func(p *NewContractParams) { p.Installments = 0 }, ErrInstallmentsNotPositive},
		{"cost <= 0", func(p *NewContractParams) { p.CostPrice = rub(t, "0.00") }, ErrCostPriceNotPositive},
		{
			"financed < installments",
			func(p *NewContractParams) {
				p.CostPrice = rub(t, "0.05")
				p.Markup = mkMarkup(t, "0.00")
				p.DownPayment = rub(t, "0.00")
				p.Installments = 6 // 5 kopecks for 6 installments
			},
			ErrFinancedLessThanInstallment,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := baseParams(t)
			tc.mutate(&p)
			_, err := NewContract(p)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestNewMarkup_NegativeRejected(t *testing.T) {
	_, err := NewMarkup(money.New(decimal.RequireFromString("-1.00"), "RUB"))
	require.ErrorIs(t, err, ErrMarkupNegative)
}

// 11 (payment guards): payment > outstanding and payment <= 0.
func TestRegisterPayment_InvalidAmounts(t *testing.T) {
	c := activeContract(t, baseParams(t)) // outstanding 90000

	err := c.RegisterPayment("p", rub(t, "90000.01"), time.Now())
	require.ErrorIs(t, err, ErrPaymentExceedsOutstanding)

	err = c.RegisterPayment("p", rub(t, "0.00"), time.Now())
	require.ErrorIs(t, err, ErrPaymentNotPositive)

	assert.Equal(t, "90000.00", c.Outstanding().String(), "rejected payments leave debt unchanged")
}
