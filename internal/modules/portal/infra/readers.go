package infra

import (
	"context"
	"errors"

	crmdomain "github.com/astralis-s/hakaton-ansar/internal/modules/crm/domain"
	financingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// ClientReader adapts the crm client repository to the portal's ClientReader port.
type ClientReader struct {
	clients crmdomain.ClientRepository
}

func NewClientReader(clients crmdomain.ClientRepository) *ClientReader {
	return &ClientReader{clients: clients}
}

var _ domain.ClientReader = (*ClientReader)(nil)

func (r *ClientReader) Get(ctx context.Context, orgID, clientID string) (domain.ClientInfo, error) {
	c, err := r.clients.GetByID(ctx, orgID, clientID)
	if err != nil {
		if errors.Is(err, crmdomain.ErrClientNotFound) {
			return domain.ClientInfo{}, domain.ErrClientNotFound
		}
		return domain.ClientInfo{}, err
	}
	return domain.ClientInfo{ID: c.ID(), FullName: c.FullName(), Phone: c.Phone()}, nil
}

func (r *ClientReader) Names(ctx context.Context, orgID string, ids []string) (map[string]string, error) {
	if len(ids) == 0 {
		return map[string]string{}, nil
	}
	all, err := r.clients.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]string, len(all))
	for _, c := range all {
		byID[c.ID()] = c.FullName()
	}
	out := make(map[string]string, len(ids))
	for _, id := range ids {
		if name, ok := byID[id]; ok {
			out[id] = name
		}
	}
	return out, nil
}

// ContractReader adapts the financing contract repository to the portal's
// ContractReader port, filtering the org's contracts down to one client's.
type ContractReader struct {
	contracts financingdomain.ContractRepository
}

func NewContractReader(contracts financingdomain.ContractRepository) *ContractReader {
	return &ContractReader{contracts: contracts}
}

var _ domain.ContractReader = (*ContractReader)(nil)

func (r *ContractReader) ListForClient(ctx context.Context, orgID, clientID string) ([]domain.ContractView, error) {
	summaries, err := r.contracts.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ContractView, 0)
	for _, s := range summaries {
		if s.ClientID != clientID {
			continue
		}
		out = append(out, domain.ContractView{
			ID:           s.ID,
			ProductID:    s.ProductID,
			SalePrice:    s.SalePrice,
			Outstanding:  s.Outstanding,
			Status:       string(s.Status),
			Installments: s.InstallmentsCount,
			CreatedAt:    s.CreatedAt,
		})
	}
	return out, nil
}
