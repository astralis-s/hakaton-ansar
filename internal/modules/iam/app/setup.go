package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// RegisterInput / RegisterOutput create a new organization with its owner.
type RegisterInput struct {
	OrgName       string
	Currency      string
	OwnerName     string
	OwnerEmail    string
	OwnerPassword string
}

type RegisterOutput struct {
	Org   domain.Organization
	Owner domain.User
}

// RegisterOrganization creates a brand-new organization and its owner atomically.
// Unlike the first-run setup it does NOT require the instance to be empty —
// Amana is multi-organization, so anyone can sign up (the owner email must be
// globally unique). This is what the public registration form calls.
type RegisterOrganization struct {
	orgs   domain.OrganizationRepository
	users  domain.UserRepository
	hasher domain.PasswordHasher
	tx     domain.TxManager
}

func NewRegisterOrganization(orgs domain.OrganizationRepository, users domain.UserRepository, hasher domain.PasswordHasher, tx domain.TxManager) *RegisterOrganization {
	return &RegisterOrganization{orgs: orgs, users: users, hasher: hasher, tx: tx}
}

func (uc *RegisterOrganization) Execute(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	hash, err := uc.hasher.Hash(in.OwnerPassword)
	if err != nil {
		return RegisterOutput{}, err
	}

	org, err := domain.NewOrganization(NewID(), in.OrgName, in.Currency)
	if err != nil {
		return RegisterOutput{}, err
	}
	owner, err := domain.NewUser(NewID(), org.ID(), in.OwnerName, in.OwnerEmail, hash, domain.RoleOwner)
	if err != nil {
		return RegisterOutput{}, err
	}

	var out RegisterOutput
	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		createdOrg, err := uc.orgs.Create(ctx, org)
		if err != nil {
			return err
		}
		createdOwner, err := uc.users.Create(ctx, owner)
		if err != nil {
			return err
		}
		out = RegisterOutput{Org: createdOrg, Owner: createdOwner}
		return nil
	})
	if err != nil {
		return RegisterOutput{}, err
	}
	return out, nil
}

// SetupOrganization is the first-run use-case (/setup): it creates the first
// organization, refusing to run once any organization already exists. It shares
// its creation logic with RegisterOrganization.
type SetupOrganization struct {
	orgs     domain.OrganizationRepository
	register *RegisterOrganization
}

func NewSetupOrganization(orgs domain.OrganizationRepository, register *RegisterOrganization) *SetupOrganization {
	return &SetupOrganization{orgs: orgs, register: register}
}

// SetupInput / SetupOutput mirror the registration types (first-run alias).
type SetupInput = RegisterInput
type SetupOutput = RegisterOutput

func (uc *SetupOrganization) Execute(ctx context.Context, in SetupInput) (SetupOutput, error) {
	count, err := uc.orgs.Count(ctx)
	if err != nil {
		return SetupOutput{}, err
	}
	if count > 0 {
		return SetupOutput{}, domain.ErrAlreadyInitialized
	}
	return uc.register.Execute(ctx, in)
}
