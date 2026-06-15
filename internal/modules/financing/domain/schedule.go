package domain

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// computeTerms validates the money inputs and derives SalePrice and
// FinancedAmount. Shared by NewContract and Preview so they agree exactly.
func computeTerms(cost money.Money, markup Markup, down money.Money) (sale, financed money.Money, err error) {
	if !cost.IsPositive() {
		return money.Money{}, money.Money{}, ErrCostPriceNotPositive
	}
	mk := markup.Money()
	if mk.Currency() != cost.Currency() || down.Currency() != cost.Currency() {
		return money.Money{}, money.Money{}, ErrCurrencyMismatch
	}

	sale, err = cost.Add(mk)
	if err != nil {
		return money.Money{}, money.Money{}, err
	}
	if down.IsNegative() {
		return money.Money{}, money.Money{}, ErrDownPaymentNegative
	}
	cmp, err := down.Cmp(sale)
	if err != nil {
		return money.Money{}, money.Money{}, err
	}
	if cmp >= 0 {
		return money.Money{}, money.Money{}, ErrDownPaymentTooLarge
	}

	financed, err = sale.Sub(down)
	if err != nil {
		return money.Money{}, money.Money{}, err
	}
	return sale, financed, nil
}

// BuildSchedule splits the financed amount into `installments` planned payments
// using deterministic rounding in minor units: every payment gets base =
// total/N, and the first `remainder` payments get one extra minor unit, so the
// payments sum to the financed amount exactly (skill: murabaha-engine rounding).
func BuildSchedule(financed money.Money, installments int, cadence Cadence, start time.Time) ([]Installment, error) {
	if installments < 1 {
		return nil, ErrInstallmentsNotPositive
	}
	if !cadence.Valid() {
		return nil, ErrInvalidCadence
	}

	total := financed.Cents()
	n := int64(installments)
	if total < n {
		// Invariant 3: each installment must be at least one minor unit.
		return nil, ErrFinancedLessThanInstallment
	}

	base := total / n
	remainder := total % n
	currency := financed.Currency()

	schedule := make([]Installment, 0, installments)
	for i := 0; i < installments; i++ {
		cents := base
		if int64(i) < remainder {
			cents++ // distribute the remainder onto the earliest payments
		}
		schedule = append(schedule, NewInstallment(i+1, cadence.addN(start, i), money.FromCents(cents, currency)))
	}
	return schedule, nil
}

// scheduleTotal sums installment amounts (used to assert invariant 4).
func scheduleTotal(schedule []Installment, currency string) (money.Money, error) {
	sum := money.Zero(currency)
	for _, inst := range schedule {
		next, err := sum.Add(inst.Amount())
		if err != nil {
			return money.Money{}, err
		}
		sum = next
	}
	return sum, nil
}

// --- Preview (pure calculation, no aggregate, no persistence) ---------------

// PreviewInput mirrors the contract terms needed to compute a schedule.
type PreviewInput struct {
	CostPrice    money.Money
	Markup       Markup
	DownPayment  money.Money
	Installments int
	Cadence      Cadence
	StartDate    time.Time
}

// PreviewResult is the computed schedule and headline figures plus the
// illustrative comparison against an interest-bearing credit.
type PreviewResult struct {
	SalePrice      money.Money
	FinancedAmount money.Money
	Schedule       []Installment
	Comparison     Comparison
}

// Preview computes the schedule exactly like NewContract but without building an
// aggregate or persisting anything. comparisonAnnualRatePercent drives only the
// illustrative riba comparison.
func Preview(in PreviewInput, comparisonAnnualRatePercent decimal.Decimal) (PreviewResult, error) {
	sale, financed, err := computeTerms(in.CostPrice, in.Markup, in.DownPayment)
	if err != nil {
		return PreviewResult{}, err
	}
	schedule, err := BuildSchedule(financed, in.Installments, in.Cadence, in.StartDate)
	if err != nil {
		return PreviewResult{}, err
	}
	return PreviewResult{
		SalePrice:      sale,
		FinancedAmount: financed,
		Schedule:       schedule,
		Comparison:     buildComparison(financed, sale, in.Installments, in.Cadence, comparisonAnnualRatePercent),
	}, nil
}

// Comparison contrasts the fixed murabaha total with a hypothetical conventional
// credit. It is illustrative (for the wizard's "no-riba vs credit" chart) and
// not a real product. Computed in decimal — no float.
type Comparison struct {
	MurabahaTotal     money.Money     // fixed: equals SalePrice
	ConventionalTotal money.Money     // hypothetical interest-bearing total
	Overpayment       money.Money     // ConventionalTotal − SalePrice (the riba you avoid)
	AnnualRatePercent decimal.Decimal // rate used for the illustration
}

func buildComparison(financed, sale money.Money, installments int, cadence Cadence, annualRatePercent decimal.Decimal) Comparison {
	rateFraction := annualRatePercent.Div(decimal.NewFromInt(100))
	years := decimal.NewFromInt(int64(installments)).Div(decimal.NewFromInt(int64(cadence.PeriodsPerYear())))
	// Simple-interest illustration: interest = financed × rate × years.
	interest := financed.Mul(rateFraction.Mul(years))
	conventional, _ := sale.Add(interest) // same currency by construction
	return Comparison{
		MurabahaTotal:     sale,
		ConventionalTotal: conventional,
		Overpayment:       interest,
		AnnualRatePercent: annualRatePercent,
	}
}
