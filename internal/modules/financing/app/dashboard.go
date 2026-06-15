package app

import (
	"context"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Dashboard aggregates the owner's morning view: overdue (with names), this
// week's expected vs collected, the portfolio balance and the upcoming payments.
// The money math lives in the domain (BuildDashboard); this use-case only loads
// data and resolves client names.
type Dashboard struct {
	contracts domain.ContractRepository
	clients   domain.ClientReader
}

func NewDashboard(contracts domain.ContractRepository, clients domain.ClientReader) *Dashboard {
	return &Dashboard{contracts: contracts, clients: clients}
}

func (uc *Dashboard) Execute(ctx context.Context, orgID string) (domain.DashboardResult, error) {
	contracts, err := uc.contracts.ListFullByOrg(ctx, orgID)
	if err != nil {
		return domain.DashboardResult{}, err
	}

	seen := make(map[string]struct{})
	ids := make([]string, 0, len(contracts))
	for _, c := range contracts {
		if _, ok := seen[c.ClientID()]; !ok {
			seen[c.ClientID()] = struct{}{}
			ids = append(ids, c.ClientID())
		}
	}
	names, err := uc.clients.Names(ctx, orgID, ids)
	if err != nil {
		return domain.DashboardResult{}, err
	}

	return domain.BuildDashboard(time.Now(), contracts, names, money.DefaultCurrency)
}
