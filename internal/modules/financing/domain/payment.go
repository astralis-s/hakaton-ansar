package domain

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Payment is a registered receipt of money against a contract.
type Payment struct {
	id     string
	amount money.Money
	paidAt time.Time
}

func NewPayment(id string, amount money.Money, paidAt time.Time) Payment {
	return Payment{id: id, amount: amount, paidAt: paidAt}
}

func (p Payment) ID() string          { return p.id }
func (p Payment) Amount() money.Money { return p.amount }
func (p Payment) PaidAt() time.Time   { return p.paidAt }
