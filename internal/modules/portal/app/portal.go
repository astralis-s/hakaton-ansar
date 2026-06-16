package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/document"
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

// GetClientContract returns the full detail (schedule + payments) of one of the
// logged-in client's contracts.
type GetClientContract struct {
	contracts domain.ContractReader
}

func NewGetClientContract(contracts domain.ContractReader) *GetClientContract {
	return &GetClientContract{contracts: contracts}
}

func (uc *GetClientContract) Execute(ctx context.Context, orgID, clientID, contractID string) (domain.ContractDetail, error) {
	return uc.contracts.GetForClient(ctx, orgID, clientID, contractID)
}

// GetContractDoc builds the printable agreement for one of the client's
// contracts (PDF rendering happens in the transport layer).
type GetContractDoc struct {
	builder domain.ContractDocBuilder
}

func NewGetContractDoc(builder domain.ContractDocBuilder) *GetContractDoc {
	return &GetContractDoc{builder: builder}
}

func (uc *GetContractDoc) Execute(ctx context.Context, orgID, contractID, clientID string) (document.Contract, error) {
	return uc.builder.Build(ctx, orgID, contractID, clientID)
}
