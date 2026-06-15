package app

import (
	"context"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// RegisterPayment records a payment (any amount ≤ outstanding) against a contract.
type RegisterPayment struct {
	contracts domain.ContractRepository
	tx        domain.TxManager
}

func NewRegisterPayment(contracts domain.ContractRepository, tx domain.TxManager) *RegisterPayment {
	return &RegisterPayment{contracts: contracts, tx: tx}
}

type RegisterPaymentInput struct {
	OrgID      string
	ContractID string
	Amount     money.Money
}

func (uc *RegisterPayment) Execute(ctx context.Context, in RegisterPaymentInput) (*domain.Contract, error) {
	var result *domain.Contract
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		contract, err := uc.contracts.GetByID(ctx, in.OrgID, in.ContractID)
		if err != nil {
			return err
		}
		if err := contract.RegisterPayment(NewID(), in.Amount, time.Now().UTC()); err != nil {
			return err
		}
		if err := uc.contracts.AddPayment(ctx, contract.ID(), lastPayment(contract)); err != nil {
			return err
		}
		if err := uc.contracts.SaveState(ctx, contract); err != nil {
			return err
		}
		result = contract
		return nil
	})
	return result, err
}

// SettleEarly pays off the whole outstanding balance with no penalty.
type SettleEarly struct {
	contracts domain.ContractRepository
	tx        domain.TxManager
}

func NewSettleEarly(contracts domain.ContractRepository, tx domain.TxManager) *SettleEarly {
	return &SettleEarly{contracts: contracts, tx: tx}
}

func (uc *SettleEarly) Execute(ctx context.Context, orgID, contractID string) (*domain.Contract, error) {
	var result *domain.Contract
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		contract, err := uc.contracts.GetByID(ctx, orgID, contractID)
		if err != nil {
			return err
		}
		if err := contract.SettleEarly(NewID(), time.Now().UTC()); err != nil {
			return err
		}
		if err := uc.contracts.AddPayment(ctx, contract.ID(), lastPayment(contract)); err != nil {
			return err
		}
		if err := uc.contracts.SaveState(ctx, contract); err != nil {
			return err
		}
		result = contract
		return nil
	})
	return result, err
}

// lastPayment returns the most recently appended payment of the aggregate.
func lastPayment(c *domain.Contract) domain.Payment {
	payments := c.Payments()
	return payments[len(payments)-1]
}
