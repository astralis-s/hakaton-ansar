package domain

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Contract is the murabaha aggregate root. The schedule, payments and
// outstanding balance can only change through its methods, which protect the
// anti-riba invariants (the debt never grows with time).
type Contract struct {
	id        string
	orgID     string
	clientID  string
	productID string

	costPrice      money.Money
	markup         Markup
	salePrice      money.Money
	downPayment    money.Money
	financedAmount money.Money
	outstanding    money.Money

	schedule []Installment
	payments []Payment

	status    ContractStatus
	cadence   Cadence
	startDate time.Time
	createdAt time.Time
}

// NewContractParams are the inputs to NewContract.
type NewContractParams struct {
	ID           string
	OrgID        string
	ClientID     string
	ProductID    string
	CostPrice    money.Money
	Markup       Markup
	DownPayment  money.Money
	Installments int
	Cadence      Cadence
	StartDate    time.Time
}

// NewContract builds a Draft contract with a fully computed schedule, enforcing
// all murabaha invariants. An invalid contract cannot be constructed.
func NewContract(p NewContractParams) (*Contract, error) {
	if p.ID == "" {
		return nil, ErrContractIDRequired
	}
	if p.OrgID == "" {
		return nil, ErrOrgIDRequired
	}
	if p.ClientID == "" {
		return nil, ErrClientIDRequired
	}
	if p.ProductID == "" {
		return nil, ErrProductIDRequired
	}

	sale, financed, err := computeTerms(p.CostPrice, p.Markup, p.DownPayment)
	if err != nil {
		return nil, err
	}

	schedule, err := BuildSchedule(financed, p.Installments, p.Cadence, p.StartDate)
	if err != nil {
		return nil, err
	}

	// Invariant 4: Σ Installment == FinancedAmount (exact).
	sum, err := scheduleTotal(schedule, financed.Currency())
	if err != nil {
		return nil, err
	}
	if !sum.Equals(financed) {
		return nil, ErrScheduleMismatch
	}
	// Invariant 5: DownPayment + Σ Installment == SalePrice.
	check, err := p.DownPayment.Add(sum)
	if err != nil {
		return nil, err
	}
	if !check.Equals(sale) {
		return nil, ErrScheduleMismatch
	}

	return &Contract{
		id:             p.ID,
		orgID:          p.OrgID,
		clientID:       p.ClientID,
		productID:      p.ProductID,
		costPrice:      p.CostPrice,
		markup:         p.Markup,
		salePrice:      sale,
		downPayment:    p.DownPayment,
		financedAmount: financed,
		outstanding:    financed, // Outstanding starts at SalePrice − DownPayment
		schedule:       schedule,
		payments:       nil,
		status:         StatusDraft,
		cadence:        p.Cadence,
		startDate:      p.StartDate,
		createdAt:      time.Now().UTC(),
	}, nil
}

// RehydrateContract rebuilds a contract from trusted storage (no recomputation).
func RehydrateContract(
	id, orgID, clientID, productID string,
	costPrice money.Money, markup Markup, salePrice, downPayment, financedAmount, outstanding money.Money,
	schedule []Installment, payments []Payment,
	status ContractStatus, cadence Cadence, startDate, createdAt time.Time,
) *Contract {
	return &Contract{
		id:             id,
		orgID:          orgID,
		clientID:       clientID,
		productID:      productID,
		costPrice:      costPrice,
		markup:         markup,
		salePrice:      salePrice,
		downPayment:    downPayment,
		financedAmount: financedAmount,
		outstanding:    outstanding,
		schedule:       schedule,
		payments:       payments,
		status:         status,
		cadence:        cadence,
		startDate:      startDate,
		createdAt:      createdAt,
	}
}

// --- State machine ----------------------------------------------------------

// Activate moves a Draft contract to Active.
func (c *Contract) Activate() error {
	if c.status != StatusDraft {
		return ErrInvalidStatusTransition
	}
	c.status = StatusActive
	return nil
}

// Cancel moves a Draft or Active contract to Cancelled (owner action).
func (c *Contract) Cancel() error {
	if c.status != StatusDraft && c.status != StatusActive {
		return ErrInvalidStatusTransition
	}
	c.status = StatusCancelled
	return nil
}

// RegisterPayment applies a payment of any amount in (0, Outstanding]. The debt
// only ever decreases; reaching zero completes the contract.
func (c *Contract) RegisterPayment(id string, amount money.Money, paidAt time.Time) error {
	if c.status != StatusActive {
		return ErrContractNotActive
	}
	if amount.Currency() != c.outstanding.Currency() {
		return ErrCurrencyMismatch
	}
	if !amount.IsPositive() {
		return ErrPaymentNotPositive
	}
	cmp, err := amount.Cmp(c.outstanding)
	if err != nil {
		return err
	}
	if cmp > 0 {
		return ErrPaymentExceedsOutstanding
	}

	newOutstanding, err := c.outstanding.Sub(amount)
	if err != nil {
		return err
	}
	c.outstanding = newOutstanding
	c.payments = append(c.payments, NewPayment(id, amount, paidAt))
	if c.outstanding.IsZero() {
		c.status = StatusCompleted
	}
	return nil
}

// SettleEarly pays the whole outstanding balance at once, with no penalty.
func (c *Contract) SettleEarly(paymentID string, paidAt time.Time) error {
	if c.status != StatusActive {
		return ErrContractNotActive
	}
	if c.outstanding.IsZero() {
		return ErrAlreadySettled
	}
	c.payments = append(c.payments, NewPayment(paymentID, c.outstanding, paidAt))
	c.outstanding = money.Zero(c.outstanding.Currency())
	c.status = StatusCompleted
	return nil
}

// --- Derived views & getters ------------------------------------------------

// PaidAmount is FinancedAmount − Outstanding.
func (c *Contract) PaidAmount() money.Money {
	paid, err := c.financedAmount.Sub(c.outstanding)
	if err != nil {
		return money.Zero(c.financedAmount.Currency())
	}
	return paid
}

// Installments returns the schedule with each installment's status derived from
// the accumulated payment as of asOf.
func (c *Contract) Installments(asOf time.Time) []InstallmentView {
	paidCents := c.PaidAmount().Cents()
	views := make([]InstallmentView, 0, len(c.schedule))
	var lowerCum int64
	for _, inst := range c.schedule {
		upperCum := lowerCum + inst.Amount().Cents()
		views = append(views, InstallmentView{
			Number:  inst.Number(),
			DueDate: inst.DueDate(),
			Amount:  inst.Amount(),
			Status:  deriveStatus(lowerCum, upperCum, paidCents, inst.DueDate(), asOf),
		})
		lowerCum = upperCum
	}
	return views
}

// HasOverdue reports whether any installment is overdue as of asOf.
func (c *Contract) HasOverdue(asOf time.Time) bool {
	for _, v := range c.Installments(asOf) {
		if v.Status == InstallmentOverdue {
			return true
		}
	}
	return false
}

func (c *Contract) ID() string                  { return c.id }
func (c *Contract) OrgID() string               { return c.orgID }
func (c *Contract) ClientID() string            { return c.clientID }
func (c *Contract) ProductID() string           { return c.productID }
func (c *Contract) CostPrice() money.Money      { return c.costPrice }
func (c *Contract) Markup() Markup              { return c.markup }
func (c *Contract) SalePrice() money.Money      { return c.salePrice }
func (c *Contract) DownPayment() money.Money    { return c.downPayment }
func (c *Contract) FinancedAmount() money.Money { return c.financedAmount }
func (c *Contract) Outstanding() money.Money    { return c.outstanding }
func (c *Contract) Schedule() []Installment     { return c.schedule }
func (c *Contract) Payments() []Payment         { return c.payments }
func (c *Contract) Status() ContractStatus      { return c.status }
func (c *Contract) Cadence() Cadence            { return c.cadence }
func (c *Contract) StartDate() time.Time        { return c.startDate }
func (c *Contract) CreatedAt() time.Time        { return c.createdAt }
func (c *Contract) InstallmentsCount() int      { return len(c.schedule) }
