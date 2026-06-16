package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// BrowseProducts lists the products a client may request (halal and in stock).
type BrowseProducts struct {
	catalog domain.CatalogReader
}

func NewBrowseProducts(catalog domain.CatalogReader) *BrowseProducts {
	return &BrowseProducts{catalog: catalog}
}

func (uc *BrowseProducts) Execute(ctx context.Context, orgID string) ([]domain.ProductCard, error) {
	return uc.catalog.ListAvailable(ctx, orgID)
}

// SubmitRequest records a client's application for a contract.
type SubmitRequest struct {
	requests domain.RequestService
}

func NewSubmitRequest(requests domain.RequestService) *SubmitRequest {
	return &SubmitRequest{requests: requests}
}

type SubmitRequestInput struct {
	OrgID               string
	ClientID            string
	ProductID           string
	DesiredInstallments int
	DesiredDownPayment  money.Money
	Note                string
}

func (uc *SubmitRequest) Execute(ctx context.Context, in SubmitRequestInput) (domain.RequestView, error) {
	return uc.requests.Submit(ctx, domain.NewRequestInput{
		OrgID:               in.OrgID,
		ClientID:            in.ClientID,
		ProductID:           in.ProductID,
		DesiredInstallments: in.DesiredInstallments,
		DesiredDownPayment:  in.DesiredDownPayment,
		Note:                in.Note,
	})
}

// ListMyRequests returns the client's own contract requests.
type ListMyRequests struct {
	requests domain.RequestService
}

func NewListMyRequests(requests domain.RequestService) *ListMyRequests {
	return &ListMyRequests{requests: requests}
}

func (uc *ListMyRequests) Execute(ctx context.Context, orgID, clientID string) ([]domain.RequestView, error) {
	return uc.requests.ListForClient(ctx, orgID, clientID)
}
