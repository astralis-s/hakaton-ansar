package domain

import (
	"context"
	"time"
)

// Location is a geographic location with its timezone (Grozny by default).
type Location struct {
	Lat float64
	Lon float64
	TZ  *time.Location
}

// PrayerTimes are the five daily prayer times for a date at a location.
type PrayerTimes struct {
	Fajr    time.Time
	Dhuhr   time.Time
	Asr     time.Time
	Maghrib time.Time
	Isha    time.Time
}

// PrayerTimesProvider yields prayer times for a date. The domain depends on this
// port, never on the prayer-time library (implemented in infra over go-prayer).
type PrayerTimesProvider interface {
	For(ctx context.Context, date time.Time, loc Location) (PrayerTimes, error)
}

// ReminderRepository persists reminders, scoped to an organization.
type ReminderRepository interface {
	Create(ctx context.Context, r Reminder) (Reminder, error)
	ListByOrg(ctx context.Context, orgID string) ([]Reminder, error)
}
