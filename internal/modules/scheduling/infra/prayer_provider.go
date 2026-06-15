// Package infra provides the scheduling adapters: the prayer-times provider over
// go-prayer and the reminder repository (sqlc).
package infra

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	goprayer "github.com/hablullah/go-prayer"

	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
)

// PrayerProvider implements domain.PrayerTimesProvider over hablullah/go-prayer.
// The domain knows nothing about this library (it depends only on the port).
// Yearly schedules are computed once and cached.
type PrayerProvider struct {
	loc      domain.Location
	asr      goprayer.AsrConvention
	twilight *goprayer.TwilightConvention

	mu    sync.Mutex
	cache map[int][]goprayer.Schedule
}

// NewPrayerProvider builds the provider for a location. madhab defaults to
// Shafi'i (Grozny); method is the twilight convention (default MWL).
func NewPrayerProvider(loc domain.Location, madhab, method string) *PrayerProvider {
	asr := goprayer.Shafii
	if strings.EqualFold(strings.TrimSpace(madhab), "hanafi") {
		asr = goprayer.Hanafi
	}
	return &PrayerProvider{
		loc:      loc,
		asr:      asr,
		twilight: twilightFor(method),
		cache:    make(map[int][]goprayer.Schedule),
	}
}

var _ domain.PrayerTimesProvider = (*PrayerProvider)(nil)

func twilightFor(method string) *goprayer.TwilightConvention {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "ISNA":
		return goprayer.ISNA()
	case "EGYPT":
		return goprayer.Egypt()
	case "KARACHI":
		return goprayer.Karachi()
	case "MWL", "":
		return goprayer.MWL()
	default:
		return goprayer.MWL()
	}
}

func (p *PrayerProvider) For(_ context.Context, date time.Time, _ domain.Location) (domain.PrayerTimes, error) {
	local := date
	if p.loc.TZ != nil {
		local = date.In(p.loc.TZ)
	}
	schedules, err := p.scheduleForYear(local.Year())
	if err != nil {
		return domain.PrayerTimes{}, err
	}
	idx := local.YearDay() - 1
	if idx < 0 || idx >= len(schedules) {
		return domain.PrayerTimes{}, fmt.Errorf("no prayer schedule for %s", local.Format("2006-01-02"))
	}
	s := schedules[idx]
	return domain.PrayerTimes{
		Fajr:    s.Fajr,
		Dhuhr:   s.Zuhr,
		Asr:     s.Asr,
		Maghrib: s.Maghrib,
		Isha:    s.Isha,
	}, nil
}

func (p *PrayerProvider) scheduleForYear(year int) ([]goprayer.Schedule, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if s, ok := p.cache[year]; ok {
		return s, nil
	}
	schedules, err := goprayer.Calculate(goprayer.Config{
		Latitude:           p.loc.Lat,
		Longitude:          p.loc.Lon,
		Timezone:           p.loc.TZ,
		TwilightConvention: p.twilight,
		AsrConvention:      p.asr,
	}, year)
	if err != nil {
		return nil, fmt.Errorf("calculate prayer times for %d: %w", year, err)
	}
	p.cache[year] = schedules
	return schedules, nil
}
