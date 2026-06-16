package domain

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// OverdueItem is one overdue contract in the dashboard's "what's on fire" list.
type OverdueItem struct {
	ContractID  string
	ClientID    string
	ClientName  string
	Outstanding money.Money
	DaysOverdue int
}

// UpcomingItem is one installment due soon (this week's cash flow).
type UpcomingItem struct {
	ContractID string
	ClientID   string
	ClientName string
	DueDate    time.Time
	Amount     money.Money
	Status     InstallmentStatus
}

// DashboardResult is the aggregated read model for the owner's dashboard. All
// money is computed in the domain (decimal) — never on the client.
type DashboardResult struct {
	PortfolioOutstanding  money.Money
	ActiveContracts       int
	WeekExpected          money.Money
	WeekCollected         money.Money
	CollectionRatePercent decimal.Decimal
	Overdue               []OverdueItem
	Upcoming              []UpcomingItem
}

const upcomingLimit = 5

// BuildDashboard aggregates the org's contracts (full aggregates) into the
// dashboard read model as of `now`. clientNames maps client id → display name.
// It is pure (deterministic) and easy to unit-test.
func BuildDashboard(now time.Time, contracts []*Contract, clientNames map[string]string, currency string) (DashboardResult, error) {
	today := dateOnly(now)
	weekStart := startOfWeek(today)
	weekEnd := weekStart.AddDate(0, 0, 7)

	portfolio := money.Zero(currency)
	weekExpected := money.Zero(currency)
	weekCollected := money.Zero(currency)
	active := 0
	var overdue []OverdueItem
	var upcoming []UpcomingItem

	for _, c := range contracts {
		if c.Status() != StatusActive {
			continue
		}
		active++
		var err error
		if portfolio, err = portfolio.Add(c.Outstanding()); err != nil {
			return DashboardResult{}, err
		}

		views := c.Installments(now)
		var earliestOverdue time.Time
		hasOverdue := false
		for _, v := range views {
			due := dateOnly(v.DueDate)
			if v.Status == InstallmentOverdue {
				if !hasOverdue || due.Before(earliestOverdue) {
					earliestOverdue = due
					hasOverdue = true
				}
			}
			if !due.Before(weekStart) && due.Before(weekEnd) {
				if weekExpected, err = weekExpected.Add(v.Amount); err != nil {
					return DashboardResult{}, err
				}
			}
			if v.Status != InstallmentPaid && v.Status != InstallmentOverdue && !due.Before(weekStart) && due.Before(weekEnd) {
				upcoming = append(upcoming, UpcomingItem{
					ContractID: c.ID(), ClientID: c.ClientID(), ClientName: clientNames[c.ClientID()],
					DueDate: v.DueDate, Amount: v.Amount, Status: v.Status,
				})
			}
		}
		if hasOverdue {
			days := daysBetween(earliestOverdue, today)
			if days < 1 {
				days = 1
			}
			overdue = append(overdue, OverdueItem{
				ContractID: c.ID(), ClientID: c.ClientID(), ClientName: clientNames[c.ClientID()],
				Outstanding: c.Outstanding(), DaysOverdue: days,
			})
		}
		for _, p := range c.Payments() {
			pa := p.PaidAt()
			if !pa.Before(weekStart) && !pa.After(now) {
				if weekCollected, err = weekCollected.Add(p.Amount()); err != nil {
					return DashboardResult{}, err
				}
			}
		}
	}

	// Loudest first: most days overdue, then largest balance.
	sort.SliceStable(overdue, func(i, j int) bool {
		if overdue[i].DaysOverdue != overdue[j].DaysOverdue {
			return overdue[i].DaysOverdue > overdue[j].DaysOverdue
		}
		return overdue[i].Outstanding.Cents() > overdue[j].Outstanding.Cents()
	})
	sort.SliceStable(upcoming, func(i, j int) bool { return upcoming[i].DueDate.Before(upcoming[j].DueDate) })
	if len(upcoming) > upcomingLimit {
		upcoming = upcoming[:upcomingLimit]
	}

	rate := decimal.Zero
	if weekExpected.IsPositive() {
		rate = weekCollected.Amount().Div(weekExpected.Amount()).Mul(decimal.NewFromInt(100)).Round(2)
	}

	return DashboardResult{
		PortfolioOutstanding:  portfolio,
		ActiveContracts:       active,
		WeekExpected:          weekExpected,
		WeekCollected:         weekCollected,
		CollectionRatePercent: rate,
		Overdue:               overdue,
		Upcoming:              upcoming,
	}, nil
}

func dateOnly(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// startOfWeek returns Monday 00:00 (UTC) of the week containing d.
func startOfWeek(d time.Time) time.Time {
	wd := int(d.Weekday()) // Sunday=0
	delta := (wd + 6) % 7  // days since Monday
	return d.AddDate(0, 0, -delta)
}

func daysBetween(from, to time.Time) int {
	return int(to.Sub(from).Hours() / 24)
}
