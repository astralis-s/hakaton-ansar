package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// GetClientProfile returns the logged-in client's own profile.
type GetClientProfile struct {
	clients domain.ClientReader
}

func NewGetClientProfile(clients domain.ClientReader) *GetClientProfile {
	return &GetClientProfile{clients: clients}
}

func (uc *GetClientProfile) Execute(ctx context.Context, orgID, clientID string) (domain.ClientInfo, error) {
	return uc.clients.Get(ctx, orgID, clientID)
}

// GetClientContracts returns the logged-in client's own installment contracts.
type GetClientContracts struct {
	contracts domain.ContractReader
}

func NewGetClientContracts(contracts domain.ContractReader) *GetClientContracts {
	return &GetClientContracts{contracts: contracts}
}

func (uc *GetClientContracts) Execute(ctx context.Context, orgID, clientID string) ([]domain.ContractView, error) {
	return uc.contracts.ListForClient(ctx, orgID, clientID)
}
