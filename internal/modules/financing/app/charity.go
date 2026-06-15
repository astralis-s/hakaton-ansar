package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// AccrueLateCharity records a fixed sadaqa charge for a contract. It is an
// owner-only action (enforced at HTTP). Crucially it does NOT change the
// contract's outstanding balance, schedule or status — the debt never grows
// (anti-riba). The charge goes to the charity registry only.
type AccrueLateCharity struct {
	contracts domain.ContractRepository
	charity   domain.CharityRepository
}

func NewAccrueLateCharity(contracts domain.ContractRepository, charity domain.CharityRepository) *AccrueLateCharity {
	return &AccrueLateCharity{contracts: contracts, charity: charity}
}

type AccrueLateCharityInput struct {
	OrgID      string
	ContractID string
	Amount     money.Money
	Note       string
	CreatedBy  string
}

func (uc *AccrueLateCharity) Execute(ctx context.Context, in AccrueLateCharityInput) (domain.CharityEntry, error) {
	contract, err := uc.contracts.GetByID(ctx, in.OrgID, in.ContractID)
	if err != nil {
		return domain.CharityEntry{}, err
	}
	entry, err := domain.NewCharityEntry(NewID(), in.OrgID, contract.ID(), contract.ClientID(), in.Amount, in.Note, in.CreatedBy)
	if err != nil {
		return domain.CharityEntry{}, err
	}
	// Intentionally no change to the contract: accruing charity never touches the debt.
	return uc.charity.Create(ctx, entry)
}

// ListCharity returns the organization's sadaqa registry.
type ListCharity struct {
	charity domain.CharityRepository
}

func NewListCharity(charity domain.CharityRepository) *ListCharity {
	return &ListCharity{charity: charity}
}

func (uc *ListCharity) Execute(ctx context.Context, orgID string) ([]domain.CharityEntry, error) {
	return uc.charity.ListByOrg(ctx, orgID)
}
