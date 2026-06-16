package domain

import (
	"strings"
	"time"
)

// ReminderType is the kind of scheduled action.
type ReminderType string

const (
	ReminderCall            ReminderType = "call"
	ReminderDelivery        ReminderType = "delivery"
	ReminderPaymentFollowUp ReminderType = "payment_followup"
)

func ParseReminderType(s string) (ReminderType, error) {
	switch ReminderType(strings.ToLower(strings.TrimSpace(s))) {
	case ReminderCall:
		return ReminderCall, nil
	case ReminderDelivery:
		return ReminderDelivery, nil
	case ReminderPaymentFollowUp:
		return ReminderPaymentFollowUp, nil
	default:
		return "", ErrInvalidReminderType
	}
}

func (t ReminderType) String() string { return string(t) }

// ReminderStatus is the stored lifecycle state of a task.
type ReminderStatus string

const (
	ReminderScheduled ReminderStatus = "scheduled"
	ReminderCompleted ReminderStatus = "completed"
	ReminderCancelled ReminderStatus = "cancelled"
	ReminderOverdue   ReminderStatus = "overdue" // derived for UI/API, never stored
)

func ParseReminderStatus(s string) (ReminderStatus, error) {
	switch ReminderStatus(strings.ToLower(strings.TrimSpace(s))) {
	case ReminderScheduled:
		return ReminderScheduled, nil
	case ReminderCompleted:
		return ReminderCompleted, nil
	case ReminderCancelled:
		return ReminderCancelled, nil
	default:
		return "", ErrInvalidReminderStatus
	}
}

func (s ReminderStatus) String() string { return string(s) }

func (s ReminderStatus) Stored() bool {
	return s == ReminderScheduled || s == ReminderCompleted || s == ReminderCancelled
}

// TimeOfDay is a wall-clock hour:minute (used for the Jummah window).
type TimeOfDay struct {
	Hour   int
	Minute int
}

// Policy configures how blocked windows are built around prayers.
type Policy struct {
	BufferBefore  time.Duration // default 0
	BufferAfter   time.Duration // default 20m
	JummahEnabled bool
	JummahStart   TimeOfDay // default 12:30
	JummahEnd     TimeOfDay // default 14:00
}

// DefaultPolicy returns the canonical Grozny policy.
func DefaultPolicy() Policy {
	return Policy{
		BufferBefore:  0,
		BufferAfter:   20 * time.Minute,
		JummahEnabled: true,
		JummahStart:   TimeOfDay{Hour: 12, Minute: 30},
		JummahEnd:     TimeOfDay{Hour: 14, Minute: 0},
	}
}

// PrayerWindow is a blocked time interval [Start, End) with a human reason.
type PrayerWindow struct {
	Start  time.Time
	End    time.Time
	Reason string
}

// ScheduledTime is the result of fitting a desired time around prayer windows.
type ScheduledTime struct {
	Time       time.Time
	WasShifted bool
	Reason     string // e.g. "перенесено из-за: намаз Магриб" / "...джума-намаз"
}
