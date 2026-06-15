package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// SetupOrganization is the first-run use-case (/setup): it atomically creates the
// organization and its first owner. It refuses to run once any organization
// already exists.
type SetupOrganization struct {
	orgs   domain.OrganizationRepository
	users  domain.UserRepository
	hasher domain.PasswordHasher
	tx     domain.TxManager
}

func NewSetupOrganization(orgs domain.OrganizationRepository, users domain.UserRepository, hasher domain.PasswordHasher, tx domain.TxManager) *SetupOrganization {
	return &SetupOrganization{orgs: orgs, users: users, hasher: hasher, tx: tx}
}

type SetupInput struct {
	OrgName       string
	Currency      string
	OwnerName     string
	OwnerEmail    string
	OwnerPassword string
}

type SetupOutput struct {
	Org   domain.Organization
	Owner domain.User
}

func (uc *SetupOrganization) Execute(ctx context.Context, in SetupInput) (SetupOutput, error) {
	count, err := uc.orgs.Count(ctx)
	if err != nil {
		return SetupOutput{}, err
	}
	if count > 0 {
		return SetupOutput{}, domain.ErrAlreadyInitialized
	}

	hash, err := uc.hasher.Hash(in.OwnerPassword)
	if err != nil {
		return SetupOutput{}, err
	}

	org, err := domain.NewOrganization(NewID(), in.OrgName, in.Currency)
	if err != nil {
		return SetupOutput{}, err
	}
	owner, err := domain.NewUser(NewID(), org.ID(), in.OwnerName, in.OwnerEmail, hash, domain.RoleOwner)
	if err != nil {
		return SetupOutput{}, err
	}

	var out SetupOutput
	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		createdOrg, err := uc.orgs.Create(ctx, org)
		if err != nil {
			return err
		}
		createdOwner, err := uc.users.Create(ctx, owner)
		if err != nil {
			return err
		}
		out = SetupOutput{Org: createdOrg, Owner: createdOwner}
		return nil
	})
	if err != nil {
		return SetupOutput{}, err
	}
	return out, nil
}
