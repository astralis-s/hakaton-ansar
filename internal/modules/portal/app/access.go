// Package app holds the portal use-cases (one type per scenario).
package app

import (
	"context"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// ProvisionAccess lets staff create/update a client's portal credentials.
type ProvisionAccess struct {
	accounts domain.AccountRepository
	clients  domain.ClientReader
	hasher   domain.Hasher
}

func NewProvisionAccess(accounts domain.AccountRepository, clients domain.ClientReader, hasher domain.Hasher) *ProvisionAccess {
	return &ProvisionAccess{accounts: accounts, clients: clients, hasher: hasher}
}

type ProvisionAccessInput struct {
	OrgID    string
	ClientID string
	Email    string
	Password string
}

func (uc *ProvisionAccess) Execute(ctx context.Context, in ProvisionAccessInput) (domain.PortalAccount, error) {
	if len(in.Password) < 8 {
		return domain.PortalAccount{}, domain.ErrPasswordTooShort
	}
	if _, err := uc.clients.Get(ctx, in.OrgID, in.ClientID); err != nil {
		return domain.PortalAccount{}, err // ErrClientNotFound
	}
	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return domain.PortalAccount{}, err
	}
	account, err := domain.NewPortalAccount(in.ClientID, in.OrgID, in.Email, hash)
	if err != nil {
		return domain.PortalAccount{}, err
	}
	if err := uc.accounts.Upsert(ctx, account); err != nil {
		return domain.PortalAccount{}, err
	}
	return account, nil
}

// GetAccess reports a client's portal access (email), if provisioned.
type GetAccess struct {
	accounts domain.AccountRepository
}

func NewGetAccess(accounts domain.AccountRepository) *GetAccess {
	return &GetAccess{accounts: accounts}
}

func (uc *GetAccess) Execute(ctx context.Context, orgID, clientID string) (domain.PortalAccount, error) {
	return uc.accounts.GetByClientID(ctx, orgID, clientID)
}

// LoginClient authenticates a client by email + password and issues a portal JWT.
type LoginClient struct {
	accounts domain.AccountRepository
	hasher   domain.Hasher
	tokens   domain.ClientTokenService
}

func NewLoginClient(accounts domain.AccountRepository, hasher domain.Hasher, tokens domain.ClientTokenService) *LoginClient {
	return &LoginClient{accounts: accounts, hasher: hasher, tokens: tokens}
}

type LoginOutput struct {
	Token     string
	ExpiresAt time.Time
	OrgID     string
	ClientID  string
}

func (uc *LoginClient) Execute(ctx context.Context, email, password string) (LoginOutput, error) {
	account, err := uc.accounts.GetByEmail(ctx, domain.NormalizeEmail(email))
	if err != nil {
		return LoginOutput{}, domain.ErrInvalidCredentials
	}
	if !uc.hasher.Compare(account.PasswordHash(), password) {
		return LoginOutput{}, domain.ErrInvalidCredentials
	}
	token, exp, err := uc.tokens.Issue(account.OrgID(), account.ClientID())
	if err != nil {
		return LoginOutput{}, err
	}
	return LoginOutput{Token: token, ExpiresAt: exp, OrgID: account.OrgID(), ClientID: account.ClientID()}, nil
}
