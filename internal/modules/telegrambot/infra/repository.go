// Package infra holds the telegrambot adapters: the Postgres user repository,
// the Telegram messenger, the cross-context adapters (crm client directory,
// portal chat inbox) and the staff-reply notifier wired into the portal chat.
package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
)

// UserRepository implements domain.UserRepository (and domain.OrgResolver) over sqlc.
type UserRepository struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

var (
	_ domain.UserRepository = (*UserRepository)(nil)
	_ domain.OrgResolver    = (*UserRepository)(nil)
)

func (r *UserRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *UserRepository) GetByChatID(ctx context.Context, chatID int64) (domain.TelegramUser, error) {
	row, err := r.q(ctx).GetTelegramUserByChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TelegramUser{}, domain.ErrUserNotFound
		}
		return domain.TelegramUser{}, fmt.Errorf("get telegram user: %w", err)
	}
	return userFromRow(row), nil
}

func (r *UserRepository) Save(ctx context.Context, u domain.TelegramUser) error {
	org, err := pgconv.UUID(u.OrgID())
	if err != nil {
		return fmt.Errorf("invalid org id: %w", err)
	}
	clientID, err := pgconv.NullableUUID(u.ClientID())
	if err != nil {
		return fmt.Errorf("invalid client id: %w", err)
	}
	_, err = r.q(ctx).UpsertTelegramUser(ctx, sqlcgen.UpsertTelegramUserParams{
		ChatID:   u.ChatID(),
		OrgID:    org,
		ClientID: clientID,
		Username: u.Username(),
		FullName: u.FullName(),
		Phone:    u.Phone(),
		State:    string(u.State()),
	})
	if err != nil {
		return fmt.Errorf("save telegram user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindChatByClient(ctx context.Context, orgID, clientID string) (int64, bool, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return 0, false, fmt.Errorf("invalid org id: %w", err)
	}
	cid, err := pgconv.UUID(clientID)
	if err != nil {
		return 0, false, nil // an invalid id simply has no Telegram chat
	}
	chatID, err := r.q(ctx).GetChatIDByClient(ctx, sqlcgen.GetChatIDByClientParams{OrgID: org, ClientID: cid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("find chat by client: %w", err)
	}
	return chatID, true, nil
}

// DefaultOrgID returns the single organization's id (the tenant this bot serves),
// or "" when no organization exists yet.
func (r *UserRepository) DefaultOrgID(ctx context.Context) (string, error) {
	id, err := r.q(ctx).GetDefaultOrgID(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("default org id: %w", err)
	}
	return pgconv.StrUUID(id), nil
}

func userFromRow(row sqlcgen.TelegramUser) domain.TelegramUser {
	return domain.RehydrateTelegramUser(
		row.ChatID,
		pgconv.StrUUID(row.OrgID),
		pgconv.StrUUID(row.ClientID),
		row.Username,
		row.FullName,
		row.Phone,
		domain.RegState(row.State),
		pgconv.TimeValue(row.CreatedAt),
		pgconv.TimeValue(row.UpdatedAt),
	)
}
