// Package app holds the crm use-cases (one type per scenario).
package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/astralis-s/hakaton-ansar/internal/modules/crm/domain"
)

// CreateClient registers a new client.
type CreateClient struct {
	clients domain.ClientRepository
}

func NewCreateClient(clients domain.ClientRepository) *CreateClient {
	return &CreateClient{clients: clients}
}

type CreateClientInput struct {
	OrgID    string
	FullName string
	Phone    string
	Document string
}

func (uc *CreateClient) Execute(ctx context.Context, in CreateClientInput) (domain.Client, error) {
	client, err := domain.NewClient(uuid.NewString(), in.OrgID, in.FullName, in.Phone, in.Document)
	if err != nil {
		return domain.Client{}, err
	}
	return uc.clients.Create(ctx, client)
}

// GetClient returns a single client within an organization.
type GetClient struct {
	clients domain.ClientRepository
}

func NewGetClient(clients domain.ClientRepository) *GetClient {
	return &GetClient{clients: clients}
}

func (uc *GetClient) Execute(ctx context.Context, orgID, id string) (domain.Client, error) {
	return uc.clients.GetByID(ctx, orgID, id)
}

// ListClients returns all clients of an organization.
type ListClients struct {
	clients domain.ClientRepository
}

func NewListClients(clients domain.ClientRepository) *ListClients {
	return &ListClients{clients: clients}
}

func (uc *ListClients) Execute(ctx context.Context, orgID string) ([]domain.Client, error) {
	return uc.clients.ListByOrg(ctx, orgID)
}

// UpdateClient edits an existing client.
type UpdateClient struct {
	clients domain.ClientRepository
}

func NewUpdateClient(clients domain.ClientRepository) *UpdateClient {
	return &UpdateClient{clients: clients}
}

type UpdateClientInput struct {
	OrgID    string
	ID       string
	FullName string
	Phone    string
	Document string
}

func (uc *UpdateClient) Execute(ctx context.Context, in UpdateClientInput) (domain.Client, error) {
	existing, err := uc.clients.GetByID(ctx, in.OrgID, in.ID)
	if err != nil {
		return domain.Client{}, err
	}
	updated, err := existing.Update(in.FullName, in.Phone, in.Document)
	if err != nil {
		return domain.Client{}, err
	}
	return uc.clients.Update(ctx, updated)
}
