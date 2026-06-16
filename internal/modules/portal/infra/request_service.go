package infra

import (
	"context"
	"errors"

	financingapp "github.com/astralis-s/hakaton-ansar/internal/modules/financing/app"
	financingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// RequestService adapts the financing contract-request use-cases to the portal's
// RequestService port and translates financing errors into portal errors so the
// portal HTTP layer maps them uniformly.
type RequestService struct {
	submit *financingapp.SubmitContractRequest
	list   *financingapp.ListClientRequests
}

func NewRequestService(submit *financingapp.SubmitContractRequest, list *financingapp.ListClientRequests) *RequestService {
	return &RequestService{submit: submit, list: list}
}

var _ domain.RequestService = (*RequestService)(nil)

func (r *RequestService) Submit(ctx context.Context, in domain.NewRequestInput) (domain.RequestView, error) {
	req, err := r.submit.Execute(ctx, financingapp.SubmitContractRequestInput{
		OrgID:               in.OrgID,
		ClientID:            in.ClientID,
		ProductID:           in.ProductID,
		DesiredInstallments: in.DesiredInstallments,
		DesiredDownPayment:  in.DesiredDownPayment,
		Note:                in.Note,
	})
	if err != nil {
		return domain.RequestView{}, mapRequestErr(err)
	}
	return requestView(req), nil
}

func (r *RequestService) ListForClient(ctx context.Context, orgID, clientID string) ([]domain.RequestView, error) {
	reqs, err := r.list.Execute(ctx, orgID, clientID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.RequestView, 0, len(reqs))
	for _, req := range reqs {
		out = append(out, requestView(req))
	}
	return out, nil
}

func requestView(req *financingdomain.ContractRequest) domain.RequestView {
	return domain.RequestView{
		ID:                  req.ID(),
		ProductID:           req.ProductID(),
		DesiredInstallments: req.DesiredInstallments(),
		DesiredDownPayment:  req.DesiredDownPayment(),
		Note:                req.Note(),
		Status:              req.Status().String(),
		ContractID:          req.ContractID(),
		CreatedAt:           req.CreatedAt(),
	}
}

func mapRequestErr(err error) error {
	switch {
	case errors.Is(err, financingdomain.ErrProductNotFound):
		return domain.ErrProductNotFound
	case errors.Is(err, financingdomain.ErrProductHaram):
		return domain.ErrProductHaram
	case errors.Is(err, financingdomain.ErrClientNotFound):
		return domain.ErrClientNotFound
	case errors.Is(err, financingdomain.ErrDesiredInstallmentsInvalid),
		errors.Is(err, financingdomain.ErrDownPaymentNegative):
		return domain.ErrInvalidRequest
	default:
		return err
	}
}
