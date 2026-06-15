package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// CreateApiKey issues a new public-API key for an organization. The raw secret is
// returned once (PlainKey) and never persisted — only its hash is stored.
type CreateApiKey struct {
	keys domain.ApiKeyRepository
}

func NewCreateApiKey(keys domain.ApiKeyRepository) *CreateApiKey {
	return &CreateApiKey{keys: keys}
}

type CreateApiKeyInput struct {
	OrgID string
	Name  string
}

type CreateApiKeyOutput struct {
	Key      domain.ApiKey
	PlainKey string // shown to the user exactly once
}

func (uc *CreateApiKey) Execute(ctx context.Context, in CreateApiKeyInput) (CreateApiKeyOutput, error) {
	raw, prefix, err := generateAPIKey()
	if err != nil {
		return CreateApiKeyOutput{}, err
	}
	keyHash := domain.HashAPIKey(raw)

	key, err := domain.NewApiKey(NewID(), in.OrgID, in.Name, prefix, keyHash)
	if err != nil {
		return CreateApiKeyOutput{}, err
	}
	created, err := uc.keys.Create(ctx, key)
	if err != nil {
		return CreateApiKeyOutput{}, err
	}
	return CreateApiKeyOutput{Key: created, PlainKey: raw}, nil
}

// ListApiKeys returns an organization's API keys (metadata only — never secrets).
type ListApiKeys struct {
	keys domain.ApiKeyRepository
}

func NewListApiKeys(keys domain.ApiKeyRepository) *ListApiKeys {
	return &ListApiKeys{keys: keys}
}

func (uc *ListApiKeys) Execute(ctx context.Context, orgID string) ([]domain.ApiKey, error) {
	return uc.keys.ListByOrg(ctx, orgID)
}

// RevokeApiKey revokes a key within its organization.
type RevokeApiKey struct {
	keys domain.ApiKeyRepository
}

func NewRevokeApiKey(keys domain.ApiKeyRepository) *RevokeApiKey {
	return &RevokeApiKey{keys: keys}
}

func (uc *RevokeApiKey) Execute(ctx context.Context, id, orgID string) (domain.ApiKey, error) {
	return uc.keys.Revoke(ctx, id, orgID)
}

// AuthenticateApiKey resolves a raw X-API-Key to its (active) organization. Used
// by the public-API middleware.
type AuthenticateApiKey struct {
	keys domain.ApiKeyRepository
}

func NewAuthenticateApiKey(keys domain.ApiKeyRepository) *AuthenticateApiKey {
	return &AuthenticateApiKey{keys: keys}
}

// Execute returns the owning organization id for a valid, non-revoked key.
func (uc *AuthenticateApiKey) Execute(ctx context.Context, rawKey string) (string, error) {
	key, err := uc.keys.GetByHash(ctx, domain.HashAPIKey(rawKey))
	if err != nil {
		return "", err
	}
	if !key.IsActive() {
		return "", domain.ErrApiKeyNotFound
	}
	return key.OrgID(), nil
}
