package domain

import (
	"context"
	"sort"
	"time"
)

// Scheduler is the domain service that fits a desired time around prayer windows
// (and Friday Jummah). It depends only on the PrayerTimesProvider port.
type Scheduler struct {
	provider PrayerTimesProvider
	policy   Policy
	loc      Location
}

func NewScheduler(provider PrayerTimesProvider, policy Policy, loc Location) *Scheduler {
	return &Scheduler{provider: provider, policy: policy, loc: loc}
}

// FindSlot returns the nearest free start time at or after `desired` such that the
// whole interval [start, start+dur] avoids every blocked window. The shift is
// always forward and the algorithm terminates (each step moves strictly forward
// to a window end; windows are finite).
func (s *Scheduler) FindSlot(ctx context.Context, desired time.Time, dur time.Duration) (ScheduledTime, error) {
	local := desired
	if s.loc.TZ != nil {
		local = desired.In(s.loc.TZ)
	}

	windows, err := s.buildWindows(ctx, local)
	if err != nil {
		return ScheduledTime{}, err
	}

	t := desired
	shifted := false
	reason := ""
	for {
		w, ok := firstBlocking(windows, t, dur)
		if !ok {
			break
		}
		t = w.End
		shifted = true
		reason = w.Reason
	}
	return ScheduledTime{Time: t, WasShifted: shifted, Reason: reason}, nil
}

// buildWindows constructs the blocked intervals for the date of `local`.
func (s *Scheduler) buildWindows(ctx context.Context, local time.Time) ([]PrayerWindow, error) {
	pt, err := s.provider.For(ctx, local, s.loc)
	if err != nil {
		return nil, err
	}

	prayers := []struct {
		name string
		t    time.Time
	}{
		{"намаз Фаджр", pt.Fajr},
		{"намаз Зухр", pt.Dhuhr},
		{"намаз Аср", pt.Asr},
		{"намаз Магриб", pt.Maghrib},
		{"намаз Иша", pt.Isha},
	}

	windows := make([]PrayerWindow, 0, len(prayers)+1)
	for _, p := range prayers {
		if p.t.IsZero() {
			continue
		}
		windows = append(windows, PrayerWindow{
			Start:  p.t.Add(-s.policy.BufferBefore),
			End:    p.t.Add(s.policy.BufferAfter),
			Reason: "перенесено из-за: " + p.name,
		})
	}

	if s.policy.JummahEnabled && local.Weekday() == time.Friday {
		y, m, d := local.Date()
		loc := local.Location()
		js := time.Date(y, m, d, s.policy.JummahStart.Hour, s.policy.JummahStart.Minute, 0, 0, loc)
		je := time.Date(y, m, d, s.policy.JummahEnd.Hour, s.policy.JummahEnd.Minute, 0, 0, loc)
		windows = append(windows, PrayerWindow{Start: js, End: je, Reason: "перенесено из-за: пятничный джума-намаз"})
	}

	sort.Slice(windows, func(i, j int) bool { return windows[i].Start.Before(windows[j].Start) })
	return windows, nil
}

// firstBlocking returns the earliest-start window that blocks the slot [t, t+dur].
func firstBlocking(windows []PrayerWindow, t time.Time, dur time.Duration) (PrayerWindow, bool) {
	for _, w := range windows {
		if blocks(w, t, dur) {
			return w, true
		}
	}
	return PrayerWindow{}, false
}

// blocks reports whether the slot starting at t with duration dur intersects the
// window [w.Start, w.End). A zero-duration slot (a call) is treated as the point
// t, blocked when w.Start <= t < w.End. A positive-duration slot is blocked on
// any overlap, so touching an edge (t == w.End or t+dur == w.Start) does not block.
func blocks(w PrayerWindow, t time.Time, dur time.Duration) bool {
	if dur == 0 {
		return !t.Before(w.Start) && t.Before(w.End)
	}
	end := t.Add(dur)
	return t.Before(w.End) && end.After(w.Start)
}
