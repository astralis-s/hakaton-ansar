package infra

import (
	"context"
	"errors"

	catalogdomain "github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	crmdomain "github.com/astralis-s/hakaton-ansar/internal/modules/crm/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
)

// ProductReader adapts the catalog product repository to financing's
// ProductReader port. When financing is extracted into its own service, this is
// the only piece that changes (in-memory call → RPC client).
type ProductReader struct {
	products catalogdomain.ProductRepository
}

func NewProductReader(products catalogdomain.ProductRepository) *ProductReader {
	return &ProductReader{products: products}
}

var _ domain.ProductReader = (*ProductReader)(nil)

func (r *ProductReader) Get(ctx context.Context, orgID, productID string) (domain.ProductInfo, error) {
	p, err := r.products.GetByID(ctx, orgID, productID)
	if err != nil {
		if errors.Is(err, catalogdomain.ErrProductNotFound) {
			return domain.ProductInfo{}, domain.ErrProductNotFound
		}
		return domain.ProductInfo{}, err
	}
	return domain.ProductInfo{ID: p.ID(), IsHaram: p.IsHaram()}, nil
}

// ClientReader adapts the crm client repository to financing's ClientReader port.
type ClientReader struct {
	clients crmdomain.ClientRepository
}

func NewClientReader(clients crmdomain.ClientRepository) *ClientReader {
	return &ClientReader{clients: clients}
}

var _ domain.ClientReader = (*ClientReader)(nil)

func (r *ClientReader) Exists(ctx context.Context, orgID, clientID string) (bool, error) {
	_, err := r.clients.GetByID(ctx, orgID, clientID)
	if err != nil {
		if errors.Is(err, crmdomain.ErrClientNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Names resolves the requested client ids to their full names with a single
// org-scoped query.
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
	want := make(map[string]string, len(ids))
	for _, id := range ids {
		if name, ok := byID[id]; ok {
			want[id] = name
		}
	}
	return want, nil
}
