// Package infra provides the crm persistence adapter (sqlc repository).
package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/crm/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/crm/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
)

// ClientRepository implements domain.ClientRepository over sqlc.
type ClientRepository struct{ pool *pgxpool.Pool }

func NewClientRepository(pool *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{pool: pool}
}

var _ domain.ClientRepository = (*ClientRepository)(nil)

func (r *ClientRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *ClientRepository) Create(ctx context.Context, c domain.Client) (domain.Client, error) {
	id, err := pgconv.UUID(c.ID())
	if err != nil {
		return domain.Client{}, fmt.Errorf("invalid client id: %w", err)
	}
	orgID, err := pgconv.UUID(c.OrgID())
	if err != nil {
		return domain.Client{}, fmt.Errorf("invalid org id: %w", err)
	}
	row, err := r.queries(ctx).CreateClient(ctx, sqlcgen.CreateClientParams{
		ID:       id,
		OrgID:    orgID,
		FullName: c.FullName(),
		Phone:    c.Phone(),
		Document: c.Document(),
	})
	if err != nil {
		return domain.Client{}, fmt.Errorf("create client: %w", err)
	}
	return clientFromRow(row), nil
}

func (r *ClientRepository) GetByID(ctx context.Context, orgID, id string) (domain.Client, error) {
	cid, err := pgconv.UUID(id)
	if err != nil {
		return domain.Client{}, domain.ErrClientNotFound
	}
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return domain.Client{}, domain.ErrClientNotFound
	}
	row, err := r.queries(ctx).GetClientByID(ctx, sqlcgen.GetClientByIDParams{ID: cid, OrgID: org})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Client{}, domain.ErrClientNotFound
		}
		return domain.Client{}, fmt.Errorf("get client: %w", err)
	}
	return clientFromRow(row), nil
}

func (r *ClientRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.Client, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.queries(ctx).ListClientsByOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	clients := make([]domain.Client, 0, len(rows))
	for _, row := range rows {
		clients = append(clients, clientFromRow(row))
	}
	return clients, nil
}

func (r *ClientRepository) Update(ctx context.Context, c domain.Client) (domain.Client, error) {
	id, err := pgconv.UUID(c.ID())
	if err != nil {
		return domain.Client{}, domain.ErrClientNotFound
	}
	org, err := pgconv.UUID(c.OrgID())
	if err != nil {
		return domain.Client{}, domain.ErrClientNotFound
	}
	row, err := r.queries(ctx).UpdateClient(ctx, sqlcgen.UpdateClientParams{
		ID:       id,
		OrgID:    org,
		FullName: c.FullName(),
		Phone:    c.Phone(),
		Document: c.Document(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Client{}, domain.ErrClientNotFound
		}
		return domain.Client{}, fmt.Errorf("update client: %w", err)
	}
	return clientFromRow(row), nil
}

func clientFromRow(c sqlcgen.Client) domain.Client {
	return domain.RehydrateClient(
		pgconv.StrUUID(c.ID),
		pgconv.StrUUID(c.OrgID),
		c.FullName,
		c.Phone,
		c.Document,
		pgconv.TimeValue(c.CreatedAt),
	)
}
