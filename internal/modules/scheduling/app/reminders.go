// Package app holds the scheduling use-cases.
package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
)

// NewID generates a new UUID string for an entity identifier.
func NewID() string { return uuid.NewString() }

// ScheduleReminder creates a reminder, shifting its time around prayer windows.
type ScheduleReminder struct {
	scheduler *domain.Scheduler
	reminders domain.ReminderRepository
}

func NewScheduleReminder(scheduler *domain.Scheduler, reminders domain.ReminderRepository) *ScheduleReminder {
	return &ScheduleReminder{scheduler: scheduler, reminders: reminders}
}

type ScheduleReminderInput struct {
	OrgID           string
	Type            string
	ClientID        string
	ContractID      string
	Note            string
	DesiredAt       time.Time
	DurationMinutes int
}

func (uc *ScheduleReminder) Execute(ctx context.Context, in ScheduleReminderInput) (domain.Reminder, error) {
	rType, err := domain.ParseReminderType(in.Type)
	if err != nil {
		return domain.Reminder{}, err
	}
	dur := time.Duration(in.DurationMinutes) * time.Minute

	slot, err := uc.scheduler.FindSlot(ctx, in.DesiredAt, dur)
	if err != nil {
		return domain.Reminder{}, err
	}

	reminder, err := domain.NewReminder(NewID(), in.OrgID, rType, in.ClientID, in.ContractID, in.Note, in.DesiredAt, dur, slot)
	if err != nil {
		return domain.Reminder{}, err
	}
	return uc.reminders.Create(ctx, reminder)
}

// PreviewSlot computes the suggested slot for a desired time without saving (used
// by the calendar to show whether/why a time would be shifted).
type PreviewSlot struct {
	scheduler *domain.Scheduler
}

func NewPreviewSlot(scheduler *domain.Scheduler) *PreviewSlot {
	return &PreviewSlot{scheduler: scheduler}
}

type PreviewSlotInput struct {
	DesiredAt       time.Time
	DurationMinutes int
}

func (uc *PreviewSlot) Execute(ctx context.Context, in PreviewSlotInput) (domain.ScheduledTime, error) {
	return uc.scheduler.FindSlot(ctx, in.DesiredAt, time.Duration(in.DurationMinutes)*time.Minute)
}

// ListReminders returns an organization's reminders.
type ListReminders struct {
	reminders domain.ReminderRepository
}

func NewListReminders(reminders domain.ReminderRepository) *ListReminders {
	return &ListReminders{reminders: reminders}
}

func (uc *ListReminders) Execute(ctx context.Context, orgID string) ([]domain.Reminder, error) {
	return uc.reminders.ListByOrg(ctx, orgID)
}

// GetReminder returns one reminder by id in an organization.
type GetReminder struct {
	reminders domain.ReminderRepository
}

func NewGetReminder(reminders domain.ReminderRepository) *GetReminder {
	return &GetReminder{reminders: reminders}
}

func (uc *GetReminder) Execute(ctx context.Context, orgID, reminderID string) (domain.Reminder, error) {
	return uc.reminders.GetByID(ctx, orgID, reminderID)
}

// UpdateReminder edits an existing scheduled task and recomputes the slot.
type UpdateReminder struct {
	scheduler *domain.Scheduler
	reminders domain.ReminderRepository
}

func NewUpdateReminder(scheduler *domain.Scheduler, reminders domain.ReminderRepository) *UpdateReminder {
	return &UpdateReminder{scheduler: scheduler, reminders: reminders}
}

type UpdateReminderInput struct {
	OrgID           string
	ReminderID      string
	Type            string
	ClientID        string
	ContractID      string
	Note            string
	DesiredAt       time.Time
	DurationMinutes int
}

func (uc *UpdateReminder) Execute(ctx context.Context, in UpdateReminderInput) (domain.Reminder, error) {
	reminder, err := uc.reminders.GetByID(ctx, in.OrgID, in.ReminderID)
	if err != nil {
		return domain.Reminder{}, err
	}
	rType, err := domain.ParseReminderType(in.Type)
	if err != nil {
		return domain.Reminder{}, err
	}
	dur := time.Duration(in.DurationMinutes) * time.Minute
	slot, err := uc.scheduler.FindSlot(ctx, in.DesiredAt, dur)
	if err != nil {
		return domain.Reminder{}, err
	}
	if err := reminder.Update(rType, in.ClientID, in.ContractID, in.Note, in.DesiredAt, dur, slot); err != nil {
		return domain.Reminder{}, err
	}
	return uc.reminders.Update(ctx, reminder)
}

// CompleteReminder closes a task as done.
type CompleteReminder struct {
	reminders domain.ReminderRepository
}

func NewCompleteReminder(reminders domain.ReminderRepository) *CompleteReminder {
	return &CompleteReminder{reminders: reminders}
}

func (uc *CompleteReminder) Execute(ctx context.Context, orgID, reminderID string) (domain.Reminder, error) {
	reminder, err := uc.reminders.GetByID(ctx, orgID, reminderID)
	if err != nil {
		return domain.Reminder{}, err
	}
	if err := reminder.Complete(time.Now()); err != nil {
		return domain.Reminder{}, err
	}
	return uc.reminders.Update(ctx, reminder)
}

// CancelReminder closes a task without completion.
type CancelReminder struct {
	reminders domain.ReminderRepository
}

func NewCancelReminder(reminders domain.ReminderRepository) *CancelReminder {
	return &CancelReminder{reminders: reminders}
}

func (uc *CancelReminder) Execute(ctx context.Context, orgID, reminderID string) (domain.Reminder, error) {
	reminder, err := uc.reminders.GetByID(ctx, orgID, reminderID)
	if err != nil {
		return domain.Reminder{}, err
	}
	if err := reminder.Cancel(time.Now()); err != nil {
		return domain.Reminder{}, err
	}
	return uc.reminders.Update(ctx, reminder)
}
