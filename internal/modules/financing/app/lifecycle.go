package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
)

// CancelContract cancels a draft/active contract (owner action; enforced at HTTP).
type CancelContract struct {
	contracts domain.ContractRepository
	tx        domain.TxManager
}

func NewCancelContract(contracts domain.ContractRepository, tx domain.TxManager) *CancelContract {
	return &CancelContract{contracts: contracts, tx: tx}
}

func (uc *CancelContract) Execute(ctx context.Context, orgID, contractID string) (*domain.Contract, error) {
	var result *domain.Contract
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		contract, err := uc.contracts.GetByID(ctx, orgID, contractID)
		if err != nil {
			return err
		}
		if err := contract.Cancel(); err != nil {
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
