package infra

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
)

// CharityRepository implements domain.CharityRepository over sqlc.
type CharityRepository struct{ pool *pgxpool.Pool }

func NewCharityRepository(pool *pgxpool.Pool) *CharityRepository {
	return &CharityRepository{pool: pool}
}

var _ domain.CharityRepository = (*CharityRepository)(nil)

func (r *CharityRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *CharityRepository) Create(ctx context.Context, e domain.CharityEntry) (domain.CharityEntry, error) {
	id, err := pgconv.UUID(e.ID())
	if err != nil {
		return domain.CharityEntry{}, fmt.Errorf("invalid charity id: %w", err)
	}
	orgID, err := pgconv.UUID(e.OrgID())
	if err != nil {
		return domain.CharityEntry{}, fmt.Errorf("invalid org id: %w", err)
	}
	contractID, err := pgconv.UUID(e.ContractID())
	if err != nil {
		return domain.CharityEntry{}, fmt.Errorf("invalid contract id: %w", err)
	}
	clientID, err := pgconv.UUID(e.ClientID())
	if err != nil {
		return domain.CharityEntry{}, fmt.Errorf("invalid client id: %w", err)
	}
	createdBy, err := pgconv.UUID(e.CreatedBy())
	if err != nil {
		return domain.CharityEntry{}, fmt.Errorf("invalid creator id: %w", err)
	}
	amount, err := pgconv.Numeric(e.Amount().Amount())
	if err != nil {
		return domain.CharityEntry{}, err
	}

	row, err := r.q(ctx).CreateCharityEntry(ctx, sqlcgen.CreateCharityEntryParams{
		ID:         id,
		OrgID:      orgID,
		ContractID: contractID,
		ClientID:   clientID,
		Amount:     amount,
		Status:     string(e.Status()),
		Note:       e.Note(),
		CreatedBy:  createdBy,
	})
	if err != nil {
		return domain.CharityEntry{}, fmt.Errorf("insert charity entry: %w", err)
	}
	return charityFromRow(row)
}

func (r *CharityRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.CharityEntry, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.q(ctx).ListCharityByOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list charity: %w", err)
	}
	entries := make([]domain.CharityEntry, 0, len(rows))
	for _, row := range rows {
		entry, err := charityFromRow(row)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func charityFromRow(row sqlcgen.CharityEntry) (domain.CharityEntry, error) {
	amount, err := moneyFrom(row.Amount, "RUB")
	if err != nil {
		return domain.CharityEntry{}, err
	}
	return domain.RehydrateCharityEntry(
		pgconv.StrUUID(row.ID),
		pgconv.StrUUID(row.OrgID),
		pgconv.StrUUID(row.ContractID),
		pgconv.StrUUID(row.ClientID),
		amount,
		domain.CharityStatus(row.Status),
		row.Note,
		pgconv.StrUUID(row.CreatedBy),
		pgconv.TimeValue(row.CreatedAt),
	), nil
}
