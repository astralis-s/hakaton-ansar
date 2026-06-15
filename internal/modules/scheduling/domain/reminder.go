package domain

import (
	"strings"
	"time"
)

// Reminder is a scheduled action (call, delivery, payment follow-up). Its final
// time is computed by the Scheduler (avoiding prayer windows); WasShifted/Reason
// explain any move for the UI.
type Reminder struct {
	id          string
	orgID       string
	rType       ReminderType
	clientID    string // optional ("")
	contractID  string // optional ("")
	note        string
	desiredAt   time.Time
	duration    time.Duration
	scheduledAt time.Time
	wasShifted  bool
	reason      string
	createdAt   time.Time
}

// NewReminder validates and creates a reminder from a computed slot.
func NewReminder(id, orgID string, rType ReminderType, clientID, contractID, note string, desiredAt time.Time, duration time.Duration, slot ScheduledTime) (Reminder, error) {
	if id == "" {
		return Reminder{}, ErrReminderIDRequired
	}
	if orgID == "" {
		return Reminder{}, ErrOrgIDRequired
	}
	if !rType.Valid() {
		return Reminder{}, ErrInvalidReminderType
	}
	return Reminder{
		id:          id,
		orgID:       orgID,
		rType:       rType,
		clientID:    strings.TrimSpace(clientID),
		contractID:  strings.TrimSpace(contractID),
		note:        strings.TrimSpace(note),
		desiredAt:   desiredAt,
		duration:    duration,
		scheduledAt: slot.Time,
		wasShifted:  slot.WasShifted,
		reason:      slot.Reason,
		createdAt:   time.Now().UTC(),
	}, nil
}

// RehydrateReminder rebuilds a reminder from trusted storage.
func RehydrateReminder(id, orgID string, rType ReminderType, clientID, contractID, note string, desiredAt time.Time, duration time.Duration, scheduledAt time.Time, wasShifted bool, reason string, createdAt time.Time) Reminder {
	return Reminder{
		id: id, orgID: orgID, rType: rType, clientID: clientID, contractID: contractID,
		note: note, desiredAt: desiredAt, duration: duration, scheduledAt: scheduledAt,
		wasShifted: wasShifted, reason: reason, createdAt: createdAt,
	}
}

func (t ReminderType) Valid() bool {
	return t == ReminderCall || t == ReminderDelivery || t == ReminderPaymentFollowUp
}

func (r Reminder) ID() string              { return r.id }
func (r Reminder) OrgID() string           { return r.orgID }
func (r Reminder) Type() ReminderType      { return r.rType }
func (r Reminder) ClientID() string        { return r.clientID }
func (r Reminder) ContractID() string      { return r.contractID }
func (r Reminder) Note() string            { return r.note }
func (r Reminder) DesiredAt() time.Time    { return r.desiredAt }
func (r Reminder) Duration() time.Duration { return r.duration }
func (r Reminder) ScheduledAt() time.Time  { return r.scheduledAt }
func (r Reminder) WasShifted() bool        { return r.wasShifted }
func (r Reminder) Reason() string          { return r.reason }
func (r Reminder) CreatedAt() time.Time    { return r.createdAt }
