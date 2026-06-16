// Package infra provides the portal persistence adapters (sqlc), the client JWT
// service, the bcrypt hasher and the cross-context readers (crm, financing).
package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
)

// AccountRepository implements domain.AccountRepository over sqlc.
type AccountRepository struct{ pool *pgxpool.Pool }

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

var _ domain.AccountRepository = (*AccountRepository)(nil)

func (r *AccountRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *AccountRepository) Upsert(ctx context.Context, a domain.PortalAccount) error {
	cid, err := pgconv.UUID(a.ClientID())
	if err != nil {
		return fmt.Errorf("invalid client id: %w", err)
	}
	org, err := pgconv.UUID(a.OrgID())
	if err != nil {
		return fmt.Errorf("invalid org id: %w", err)
	}
	err = r.q(ctx).UpsertPortalAccount(ctx, sqlcgen.UpsertPortalAccountParams{
		ClientID:     cid,
		OrgID:        org,
		Email:        a.Email(),
		PasswordHash: a.PasswordHash(),
	})
	if err != nil {
		if pgconv.IsUniqueViolation(err, "") {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("upsert portal account: %w", err)
	}
	return nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (domain.PortalAccount, error) {
	row, err := r.q(ctx).GetPortalAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PortalAccount{}, domain.ErrAccountNotFound
		}
		return domain.PortalAccount{}, fmt.Errorf("get account by email: %w", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) GetByClientID(ctx context.Context, orgID, clientID string) (domain.PortalAccount, error) {
	cid, err := pgconv.UUID(clientID)
	if err != nil {
		return domain.PortalAccount{}, domain.ErrAccountNotFound
	}
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return domain.PortalAccount{}, domain.ErrAccountNotFound
	}
	row, err := r.q(ctx).GetPortalAccountByClientID(ctx, sqlcgen.GetPortalAccountByClientIDParams{OrgID: org, ClientID: cid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PortalAccount{}, domain.ErrAccountNotFound
		}
		return domain.PortalAccount{}, fmt.Errorf("get account by client: %w", err)
	}
	return accountFromRow(row), nil
}

func accountFromRow(a sqlcgen.ClientPortalAccount) domain.PortalAccount {
	return domain.RehydratePortalAccount(
		pgconv.StrUUID(a.ClientID),
		pgconv.StrUUID(a.OrgID),
		a.Email,
		a.PasswordHash,
		pgconv.TimeValue(a.CreatedAt),
	)
}
