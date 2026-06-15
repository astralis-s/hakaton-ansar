package domain

import "context"

// ClientRepository persists clients, scoped to an organization.
type ClientRepository interface {
	Create(ctx context.Context, c Client) (Client, error)
	GetByID(ctx context.Context, orgID, id string) (Client, error)
	ListByOrg(ctx context.Context, orgID string) ([]Client, error)
	Update(ctx context.Context, c Client) (Client, error)
}
