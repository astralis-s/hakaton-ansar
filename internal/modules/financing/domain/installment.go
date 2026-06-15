package domain

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Installment is one planned payment in the schedule: a sequence number, a due
// date and an amount. Its status is NOT stored here — it is derived from the
// contract's accumulated payment (see InstallmentView).
type Installment struct {
	number  int
	dueDate time.Time
	amount  money.Money
}

func NewInstallment(number int, dueDate time.Time, amount money.Money) Installment {
	return Installment{number: number, dueDate: dueDate, amount: amount}
}

func (i Installment) Number() int         { return i.number }
func (i Installment) DueDate() time.Time  { return i.dueDate }
func (i Installment) Amount() money.Money { return i.amount }

// InstallmentView is an installment plus its derived status at a point in time.
type InstallmentView struct {
	Number  int
	DueDate time.Time
	Amount  money.Money
	Status  InstallmentStatus
}

// deriveStatus computes an installment's status from the accumulated paid amount
// (in minor units) and the cumulative lower/upper bounds of this installment.
//
//	Paid          — paid >= upper          (fully covered)
//	PartiallyPaid — lower < paid < upper    (exactly one such installment)
//	Overdue       — paid <= lower and due date has passed (uncovered, late)
//	Pending       — paid <= lower and due date is in the future
func deriveStatus(lowerCum, upperCum, paidCents int64, dueDate, asOf time.Time) InstallmentStatus {
	switch {
	case paidCents >= upperCum:
		return InstallmentPaid
	case paidCents > lowerCum:
		return InstallmentPartiallyPaid
	case !asOf.Before(dueDate):
		return InstallmentOverdue
	default:
		return InstallmentPending
	}
}
