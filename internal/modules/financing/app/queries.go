package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
)

// GetContract loads a full contract aggregate.
type GetContract struct {
	contracts domain.ContractRepository
}

func NewGetContract(contracts domain.ContractRepository) *GetContract {
	return &GetContract{contracts: contracts}
}

func (uc *GetContract) Execute(ctx context.Context, orgID, id string) (*domain.Contract, error) {
	return uc.contracts.GetByID(ctx, orgID, id)
}

// ListContracts returns contract summaries for an organization.
type ListContracts struct {
	contracts domain.ContractRepository
}

func NewListContracts(contracts domain.ContractRepository) *ListContracts {
	return &ListContracts{contracts: contracts}
}

func (uc *ListContracts) Execute(ctx context.Context, orgID string) ([]domain.ContractSummary, error) {
	return uc.contracts.ListByOrg(ctx, orgID)
}
