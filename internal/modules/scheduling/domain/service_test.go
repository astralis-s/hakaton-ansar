package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var msk = mustTZ()

func mustTZ() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*3600)
	}
	return loc
}

// stubProvider returns fixed prayer times (clock h:m on the requested date),
// making tests deterministic and independent of real astronomical calculation.
type stubProvider struct {
	fajr, dhuhr, asr, maghrib, isha [2]int // {hour, minute}
}

func (s stubProvider) For(_ context.Context, date time.Time, loc Location) (PrayerTimes, error) {
	at := func(hm [2]int) time.Time {
		y, m, d := date.In(loc.TZ).Date()
		return time.Date(y, m, d, hm[0], hm[1], 0, 0, loc.TZ)
	}
	return PrayerTimes{
		Fajr:    at(s.fajr),
		Dhuhr:   at(s.dhuhr),
		Asr:     at(s.asr),
		Maghrib: at(s.maghrib),
		Isha:    at(s.isha),
	}, nil
}

func at(t *testing.T, weekday time.Weekday, hour, min int) time.Time {
	t.Helper()
	// 2026-06-15 is a Monday; +4 days = Friday 2026-06-19.
	day := 15 + int(weekday-time.Monday)
	return time.Date(2026, 6, day, hour, min, 0, 0, msk)
}

func newScheduler(p PrayerTimesProvider) *Scheduler {
	policy := DefaultPolicy() // BufferAfter 20m, Jummah 12:30–14:00
	return NewScheduler(p, policy, Location{Lat: 43.3178, Lon: 45.6949, TZ: msk})
}

// standard day: Fajr 04:00, Dhuhr 12:00, Asr 16:00, Maghrib 20:00, Isha 22:00.
func standardDay() stubProvider {
	return stubProvider{fajr: [2]int{4, 0}, dhuhr: [2]int{12, 0}, asr: [2]int{16, 0}, maghrib: [2]int{20, 0}, isha: [2]int{22, 0}}
}

func TestFindSlot_OutsideWindows(t *testing.T) {
	s := newScheduler(standardDay())
	desired := at(t, time.Monday, 10, 0) // nothing near 10:00
	got, err := s.FindSlot(context.Background(), desired, 0)
	require.NoError(t, err)
	assert.False(t, got.WasShifted)
	assert.True(t, got.Time.Equal(desired))
}

func TestFindSlot_InsidePrayerWindow(t *testing.T) {
	s := newScheduler(standardDay())
	desired := at(t, time.Monday, 16, 5) // inside Asr [16:00,16:20]
	got, err := s.FindSlot(context.Background(), desired, 0)
	require.NoError(t, err)
	assert.True(t, got.WasShifted)
	assert.True(t, got.Time.Equal(at(t, time.Monday, 16, 20)), "shift to Asr+20m")
	assert.Contains(t, got.Reason, "Аср")
}

func TestFindSlot_DeliveryDurationStraddlesWindow(t *testing.T) {
	s := newScheduler(standardDay())
	desired := at(t, time.Monday, 15, 50) // [15:50,16:20] overlaps Asr [16:00,16:20]
	got, err := s.FindSlot(context.Background(), desired, 30*time.Minute)
	require.NoError(t, err)
	assert.True(t, got.WasShifted)
	assert.True(t, got.Time.Equal(at(t, time.Monday, 16, 20)), "whole interval clears the window")
}

func TestFindSlot_CascadeShift(t *testing.T) {
	// Two near-adjacent windows: Asr [16:00,16:20], Maghrib [16:10,16:30].
	p := stubProvider{fajr: [2]int{4, 0}, dhuhr: [2]int{12, 0}, asr: [2]int{16, 0}, maghrib: [2]int{16, 10}, isha: [2]int{22, 0}}
	s := newScheduler(p)
	desired := at(t, time.Monday, 15, 55) // [15:55,16:25] hits Asr → 16:20 → hits Maghrib → 16:30
	got, err := s.FindSlot(context.Background(), desired, 30*time.Minute)
	require.NoError(t, err)
	assert.True(t, got.WasShifted)
	assert.True(t, got.Time.Equal(at(t, time.Monday, 16, 30)), "cascaded past both windows")
	assert.Contains(t, got.Reason, "Магриб")
}

func TestFindSlot_FridayJummah(t *testing.T) {
	s := newScheduler(standardDay())
	desired := at(t, time.Friday, 12, 45) // inside Jummah [12:30,14:00]
	got, err := s.FindSlot(context.Background(), desired, 0)
	require.NoError(t, err)
	assert.True(t, got.WasShifted)
	assert.True(t, got.Time.Equal(at(t, time.Friday, 14, 0)), "shift past Jummah end")
	assert.Contains(t, got.Reason, "джума")
}

func TestFindSlot_NonFridayNoJummah(t *testing.T) {
	s := newScheduler(standardDay())
	desired := at(t, time.Monday, 13, 0) // would be inside Jummah window, but it's Monday
	got, err := s.FindSlot(context.Background(), desired, 0)
	require.NoError(t, err)
	assert.False(t, got.WasShifted)
	assert.True(t, got.Time.Equal(desired))
}

func TestFindSlot_BoundaryTouchNoFalseShift(t *testing.T) {
	s := newScheduler(standardDay())
	desired := at(t, time.Monday, 16, 20) // exactly Asr window End → free
	got, err := s.FindSlot(context.Background(), desired, 30*time.Minute)
	require.NoError(t, err)
	assert.False(t, got.WasShifted, "touching the window end must not shift")
	assert.True(t, got.Time.Equal(desired))
}
