package infra

import (
	"context"

	financingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
)

// SalesReader adapts the financing contract repository to the ledger's SalesReader
// port. It reads the full aggregates (so it sees cost price and status) and maps
// each contract to a Sale — the only data the income/expense report needs. When
// financing becomes its own service, only this adapter changes.
type SalesReader struct {
	contracts financingdomain.ContractRepository
}

func NewSalesReader(contracts financingdomain.ContractRepository) *SalesReader {
	return &SalesReader{contracts: contracts}
}

var _ domain.SalesReader = (*SalesReader)(nil)

func (r *SalesReader) ListSales(ctx context.Context, orgID string) ([]domain.Sale, error) {
	contracts, err := r.contracts.ListFullByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Sale, 0, len(contracts))
	for _, c := range contracts {
		out = append(out, domain.Sale{
			ContractID: c.ID(),
			ProductID:  c.ProductID(),
			SalePrice:  c.SalePrice(),
			CostPrice:  c.CostPrice(),
			Status:     string(c.Status()),
			CreatedAt:  c.CreatedAt(),
		})
	}
	return out, nil
}
