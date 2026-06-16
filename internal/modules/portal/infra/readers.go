package infra

import (
	"context"
	"errors"
	"time"

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

func (r *ContractReader) GetForClient(ctx context.Context, orgID, clientID, contractID string) (domain.ContractDetail, error) {
	c, err := r.contracts.GetByID(ctx, orgID, contractID)
	if err != nil {
		if errors.Is(err, financingdomain.ErrContractNotFound) {
			return domain.ContractDetail{}, domain.ErrContractNotFound
		}
		return domain.ContractDetail{}, err
	}
	// Authorization: a client may only see their own contract; otherwise behave
	// as if it does not exist (do not leak other clients' contract ids).
	if c.ClientID() != clientID {
		return domain.ContractDetail{}, domain.ErrContractNotFound
	}

	asOf := time.Now()
	views := c.Installments(asOf)
	lines := make([]domain.InstallmentLine, 0, len(views))
	detail := domain.ContractDetail{
		ID:             c.ID(),
		ProductID:      c.ProductID(),
		SalePrice:      c.SalePrice(),
		DownPayment:    c.DownPayment(),
		FinancedAmount: c.FinancedAmount(),
		Outstanding:    c.Outstanding(),
		PaidAmount:     c.PaidAmount(),
		Status:         string(c.Status()),
		Cadence:        c.Cadence().String(),
		StartDate:      c.StartDate(),
		HasOverdue:     c.HasOverdue(asOf),
		CreatedAt:      c.CreatedAt(),
	}
	for _, v := range views {
		lines = append(lines, domain.InstallmentLine{
			Number:  v.Number,
			DueDate: v.DueDate,
			Amount:  v.Amount,
			Status:  string(v.Status),
		})
		// The next payment is the first line that is not fully paid.
		if !detail.HasNext && v.Status != financingdomain.InstallmentPaid {
			detail.HasNext = true
			detail.NextDueDate = v.DueDate
			detail.NextDueAmount = v.Amount
		}
	}
	detail.Installments = lines

	for _, p := range c.Payments() {
		detail.Payments = append(detail.Payments, domain.PaymentLine{Amount: p.Amount(), PaidAt: p.PaidAt()})
	}
	return detail, nil
}
