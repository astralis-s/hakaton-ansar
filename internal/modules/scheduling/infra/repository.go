package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
)

// ReminderRepository implements domain.ReminderRepository over sqlc.
type ReminderRepository struct{ pool *pgxpool.Pool }

func NewReminderRepository(pool *pgxpool.Pool) *ReminderRepository {
	return &ReminderRepository{pool: pool}
}

var _ domain.ReminderRepository = (*ReminderRepository)(nil)

func (r *ReminderRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *ReminderRepository) Create(ctx context.Context, rem domain.Reminder) (domain.Reminder, error) {
	id, err := pgconv.UUID(rem.ID())
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("invalid reminder id: %w", err)
	}
	orgID, err := pgconv.UUID(rem.OrgID())
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("invalid org id: %w", err)
	}
	clientID, err := pgconv.NullableUUID(rem.ClientID())
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("invalid client id: %w", err)
	}
	contractID, err := pgconv.NullableUUID(rem.ContractID())
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("invalid contract id: %w", err)
	}

	row, err := r.q(ctx).CreateReminder(ctx, sqlcgen.CreateReminderParams{
		ID:              id,
		OrgID:           orgID,
		Type:            rem.Type().String(),
		ClientID:        clientID,
		ContractID:      contractID,
		Note:            rem.Note(),
		DesiredAt:       pgconv.Timestamp(rem.DesiredAt()),
		DurationMinutes: int32(rem.Duration() / time.Minute),
		ScheduledAt:     pgconv.Timestamp(rem.ScheduledAt()),
		WasShifted:      rem.WasShifted(),
		Reason:          rem.Reason(),
	})
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("create reminder: %w", err)
	}
	return reminderFromRow(row), nil
}

func (r *ReminderRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.Reminder, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.q(ctx).ListRemindersByOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list reminders: %w", err)
	}
	reminders := make([]domain.Reminder, 0, len(rows))
	for _, row := range rows {
		reminders = append(reminders, reminderFromRow(row))
	}
	return reminders, nil
}

func reminderFromRow(row sqlcgen.Reminder) domain.Reminder {
	return domain.RehydrateReminder(
		pgconv.StrUUID(row.ID),
		pgconv.StrUUID(row.OrgID),
		domain.ReminderType(row.Type),
		pgconv.StrUUID(row.ClientID),
		pgconv.StrUUID(row.ContractID),
		row.Note,
		pgconv.TimeValue(row.DesiredAt),
		time.Duration(row.DurationMinutes)*time.Minute,
		pgconv.TimeValue(row.ScheduledAt),
		row.WasShifted,
		row.Reason,
		pgconv.TimeValue(row.CreatedAt),
	)
}
