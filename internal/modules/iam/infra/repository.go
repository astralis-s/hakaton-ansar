// Package infra provides the IAM persistence adapters (sqlc repositories) and
// external-dependency adapters (bcrypt password hasher, JWT token service).
package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
)

// queries returns a sqlc Queries bound to the active transaction (if any) or the
// pool, via the context.
func queries(ctx context.Context, pool *pgxpool.Pool) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, pool))
}

// --- OrganizationRepository -------------------------------------------------

type OrganizationRepository struct{ pool *pgxpool.Pool }

func NewOrganizationRepository(pool *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{pool: pool}
}

var _ domain.OrganizationRepository = (*OrganizationRepository)(nil)

func (r *OrganizationRepository) Create(ctx context.Context, org domain.Organization) (domain.Organization, error) {
	id, err := pgUUID(org.ID())
	if err != nil {
		return domain.Organization{}, fmt.Errorf("invalid org id: %w", err)
	}
	row, err := queries(ctx, r.pool).CreateOrganization(ctx, sqlcgen.CreateOrganizationParams{
		ID:       id,
		Name:     org.Name(),
		Currency: org.Currency(),
	})
	if err != nil {
		return domain.Organization{}, fmt.Errorf("create organization: %w", err)
	}
	return orgFromRow(row), nil
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (domain.Organization, error) {
	uid, err := pgUUID(id)
	if err != nil {
		return domain.Organization{}, domain.ErrOrgNotFound
	}
	row, err := queries(ctx, r.pool).GetOrganizationByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Organization{}, domain.ErrOrgNotFound
		}
		return domain.Organization{}, fmt.Errorf("get organization: %w", err)
	}
	return orgFromRow(row), nil
}

func (r *OrganizationRepository) Count(ctx context.Context) (int64, error) {
	n, err := queries(ctx, r.pool).CountOrganizations(ctx)
	if err != nil {
		return 0, fmt.Errorf("count organizations: %w", err)
	}
	return n, nil
}

func orgFromRow(o sqlcgen.Organization) domain.Organization {
	return domain.RehydrateOrganization(strUUID(o.ID), o.Name, o.Currency, timeValue(o.CreatedAt))
}

// --- UserRepository ---------------------------------------------------------

type UserRepository struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

var _ domain.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u domain.User) (domain.User, error) {
	id, err := pgUUID(u.ID())
	if err != nil {
		return domain.User{}, fmt.Errorf("invalid user id: %w", err)
	}
	orgID, err := pgUUID(u.OrgID())
	if err != nil {
		return domain.User{}, fmt.Errorf("invalid org id: %w", err)
	}
	row, err := queries(ctx, r.pool).CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:           id,
		OrgID:        orgID,
		FullName:     u.FullName(),
		Email:        u.Email(),
		PasswordHash: u.PasswordHash(),
		Role:         u.Role().String(),
	})
	if err != nil {
		if isUniqueViolation(err, "users_email_unique") {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return userFromRow(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := queries(ctx, r.pool).GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return userFromRow(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	uid, err := pgUUID(id)
	if err != nil {
		return domain.User{}, domain.ErrUserNotFound
	}
	row, err := queries(ctx, r.pool).GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return userFromRow(row), nil
}

func (r *UserRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.User, error) {
	uid, err := pgUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := queries(ctx, r.pool).ListUsersByOrg(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, userFromRow(row))
	}
	return users, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	n, err := queries(ctx, r.pool).CountUsersByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("count users by email: %w", err)
	}
	return n > 0, nil
}

func userFromRow(u sqlcgen.User) domain.User {
	return domain.RehydrateUser(
		strUUID(u.ID),
		strUUID(u.OrgID),
		u.FullName,
		u.Email,
		u.PasswordHash,
		domain.Role(u.Role),
		timeValue(u.CreatedAt),
	)
}

// --- ApiKeyRepository -------------------------------------------------------

type ApiKeyRepository struct{ pool *pgxpool.Pool }

func NewApiKeyRepository(pool *pgxpool.Pool) *ApiKeyRepository {
	return &ApiKeyRepository{pool: pool}
}

var _ domain.ApiKeyRepository = (*ApiKeyRepository)(nil)

func (r *ApiKeyRepository) Create(ctx context.Context, key domain.ApiKey) (domain.ApiKey, error) {
	id, err := pgUUID(key.ID())
	if err != nil {
		return domain.ApiKey{}, fmt.Errorf("invalid api key id: %w", err)
	}
	orgID, err := pgUUID(key.OrgID())
	if err != nil {
		return domain.ApiKey{}, fmt.Errorf("invalid org id: %w", err)
	}
	row, err := queries(ctx, r.pool).CreateApiKey(ctx, sqlcgen.CreateApiKeyParams{
		ID:      id,
		OrgID:   orgID,
		Name:    key.Name(),
		Prefix:  key.Prefix(),
		KeyHash: key.KeyHash(),
	})
	if err != nil {
		return domain.ApiKey{}, fmt.Errorf("create api key: %w", err)
	}
	return apiKeyFromRow(row), nil
}

func (r *ApiKeyRepository) GetByHash(ctx context.Context, keyHash string) (domain.ApiKey, error) {
	row, err := queries(ctx, r.pool).GetApiKeyByHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ApiKey{}, domain.ErrApiKeyNotFound
		}
		return domain.ApiKey{}, fmt.Errorf("get api key by hash: %w", err)
	}
	return apiKeyFromRow(row), nil
}

func (r *ApiKeyRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.ApiKey, error) {
	uid, err := pgUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := queries(ctx, r.pool).ListApiKeysByOrg(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	keys := make([]domain.ApiKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, apiKeyFromRow(row))
	}
	return keys, nil
}

func (r *ApiKeyRepository) Revoke(ctx context.Context, id, orgID string) (domain.ApiKey, error) {
	keyID, err := pgUUID(id)
	if err != nil {
		return domain.ApiKey{}, domain.ErrApiKeyNotFound
	}
	org, err := pgUUID(orgID)
	if err != nil {
		return domain.ApiKey{}, domain.ErrApiKeyNotFound
	}
	row, err := queries(ctx, r.pool).RevokeApiKey(ctx, sqlcgen.RevokeApiKeyParams{
		ID:    keyID,
		OrgID: org,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ApiKey{}, domain.ErrApiKeyNotFound
		}
		return domain.ApiKey{}, fmt.Errorf("revoke api key: %w", err)
	}
	return apiKeyFromRow(row), nil
}

func apiKeyFromRow(k sqlcgen.ApiKey) domain.ApiKey {
	return domain.RehydrateApiKey(
		strUUID(k.ID),
		strUUID(k.OrgID),
		k.Name,
		k.Prefix,
		k.KeyHash,
		timeValue(k.CreatedAt),
		timePtr(k.RevokedAt),
	)
}
