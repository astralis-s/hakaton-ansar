package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
		Status:          rem.Status().String(),
	})
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("create reminder: %w", err)
	}
	return reminderFromRow(row), nil
}

func (r *ReminderRepository) GetByID(ctx context.Context, orgID, reminderID string) (domain.Reminder, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("invalid org id: %w", err)
	}
	id, err := pgconv.UUID(reminderID)
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("invalid reminder id: %w", err)
	}
	row, err := r.q(ctx).GetReminderByID(ctx, org, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Reminder{}, domain.ErrReminderNotFound
		}
		return domain.Reminder{}, fmt.Errorf("get reminder: %w", err)
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

func (r *ReminderRepository) Update(ctx context.Context, rem domain.Reminder) (domain.Reminder, error) {
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

	row, err := r.q(ctx).UpdateReminder(ctx, sqlcgen.UpdateReminderParams{
		OrgID:           orgID,
		ID:              id,
		Type:            rem.Type().String(),
		ClientID:        clientID,
		ContractID:      contractID,
		Note:            rem.Note(),
		DesiredAt:       pgconv.Timestamp(rem.DesiredAt()),
		DurationMinutes: int32(rem.Duration() / time.Minute),
		ScheduledAt:     pgconv.Timestamp(rem.ScheduledAt()),
		WasShifted:      rem.WasShifted(),
		Reason:          rem.Reason(),
		Status:          rem.Status().String(),
		CompletedAt:     pgconv.NullableTimestamp(rem.CompletedAt()),
		CancelledAt:     pgconv.NullableTimestamp(rem.CancelledAt()),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Reminder{}, domain.ErrReminderNotFound
		}
		return domain.Reminder{}, fmt.Errorf("update reminder: %w", err)
	}
	return reminderFromRow(row), nil
}

func reminderFromRow(row sqlcgen.Reminder) domain.Reminder {
	status, err := domain.ParseReminderStatus(row.Status)
	if err != nil {
		status = domain.ReminderScheduled
	}
	return domain.RehydrateReminder(
		pgconv.StrUUID(row.ID),
		pgconv.StrUUID(row.OrgID),
		domain.ReminderType(row.Type),
		status,
		pgconv.StrUUID(row.ClientID),
		pgconv.StrUUID(row.ContractID),
		row.Note,
		pgconv.TimeValue(row.DesiredAt),
		time.Duration(row.DurationMinutes)*time.Minute,
		pgconv.TimeValue(row.ScheduledAt),
		row.WasShifted,
		row.Reason,
		pgconv.TimePtr(row.CompletedAt),
		pgconv.TimePtr(row.CancelledAt),
		pgconv.TimeValue(row.CreatedAt),
	)
}
