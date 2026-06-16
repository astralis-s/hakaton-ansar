package infra

import (
	"context"

	catalogdomain "github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// CatalogReader adapts the catalog product repository to the portal's
// CatalogReader port, exposing only products a client may actually request
// (halal and in stock — i.e. financeable).
type CatalogReader struct {
	products catalogdomain.ProductRepository
}

func NewCatalogReader(products catalogdomain.ProductRepository) *CatalogReader {
	return &CatalogReader{products: products}
}

var _ domain.CatalogReader = (*CatalogReader)(nil)

func (r *CatalogReader) ListAvailable(ctx context.Context, orgID string) ([]domain.ProductCard, error) {
	products, err := r.products.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProductCard, 0, len(products))
	for _, p := range products {
		if !p.CanBeFinanced() {
			continue
		}
		out = append(out, domain.ProductCard{ID: p.ID(), Name: p.Name(), Category: p.Category()})
	}
	return out, nil
}
