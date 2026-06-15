package http

import (
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
)

type dashboardResponse struct {
	Portfolio portfolioDTO `json:"portfolio"`
	Week      weekDTO      `json:"week"`
	Overdue   []overdueDTO `json:"overdue"`
}

type portfolioDTO struct {
	Outstanding     string `json:"outstanding"`
	ActiveContracts int    `json:"active_contracts"`
}

type weekDTO struct {
	Expected              string        `json:"expected"`
	Collected             string        `json:"collected"`
	CollectionRatePercent string        `json:"collection_rate_percent"`
	Upcoming              []upcomingDTO `json:"upcoming"`
}

type overdueDTO struct {
	ContractID  string `json:"contract_id"`
	ClientID    string `json:"client_id"`
	ClientName  string `json:"client_name"`
	Outstanding string `json:"outstanding"`
	DaysOverdue int    `json:"days_overdue"`
}

type upcomingDTO struct {
	ContractID string `json:"contract_id"`
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
	DueDate    string `json:"due_date"`
	Amount     string `json:"amount"`
	Status     string `json:"status"`
}

func toDashboardResponse(d domain.DashboardResult) dashboardResponse {
	overdue := make([]overdueDTO, 0, len(d.Overdue))
	for _, o := range d.Overdue {
		overdue = append(overdue, overdueDTO{
			ContractID: o.ContractID, ClientID: o.ClientID, ClientName: o.ClientName,
			Outstanding: o.Outstanding.String(), DaysOverdue: o.DaysOverdue,
		})
	}
	upcoming := make([]upcomingDTO, 0, len(d.Upcoming))
	for _, u := range d.Upcoming {
		upcoming = append(upcoming, upcomingDTO{
			ContractID: u.ContractID, ClientID: u.ClientID, ClientName: u.ClientName,
			DueDate: u.DueDate.Format(dateLayout), Amount: u.Amount.String(), Status: u.Status.String(),
		})
	}
	return dashboardResponse{
		Portfolio: portfolioDTO{Outstanding: d.PortfolioOutstanding.String(), ActiveContracts: d.ActiveContracts},
		Week: weekDTO{
			Expected:              d.WeekExpected.String(),
			Collected:             d.WeekCollected.String(),
			CollectionRatePercent: d.CollectionRatePercent.String(),
			Upcoming:              upcoming,
		},
		Overdue: overdue,
	}
}
