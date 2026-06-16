package app

import (
	"context"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// SubmitContractRequest records a client's application for a contract. It refuses
// a haram product and a non-existent client/product, but does NOT set financial
// terms — those are the manager's at approval time.
type SubmitContractRequest struct {
	requests domain.ContractRequestRepository
	products domain.ProductReader
	clients  domain.ClientReader
}

func NewSubmitContractRequest(requests domain.ContractRequestRepository, products domain.ProductReader, clients domain.ClientReader) *SubmitContractRequest {
	return &SubmitContractRequest{requests: requests, products: products, clients: clients}
}

type SubmitContractRequestInput struct {
	OrgID               string
	ClientID            string
	ProductID           string
	DesiredInstallments int
	DesiredDownPayment  money.Money
	Note                string
}

func (uc *SubmitContractRequest) Execute(ctx context.Context, in SubmitContractRequestInput) (*domain.ContractRequest, error) {
	product, err := uc.products.Get(ctx, in.OrgID, in.ProductID)
	if err != nil {
		return nil, err
	}
	if product.IsHaram {
		return nil, domain.ErrProductHaram
	}
	exists, err := uc.clients.Exists(ctx, in.OrgID, in.ClientID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrClientNotFound
	}

	request, err := domain.NewContractRequest(NewID(), in.OrgID, in.ClientID, in.ProductID, in.DesiredInstallments, in.DesiredDownPayment, in.Note)
	if err != nil {
		return nil, err
	}
	if err := uc.requests.Create(ctx, request); err != nil {
		return nil, err
	}
	return request, nil
}

// ListContractRequests returns the organization's requests (staff inbox).
type ListContractRequests struct {
	requests domain.ContractRequestRepository
}

func NewListContractRequests(requests domain.ContractRequestRepository) *ListContractRequests {
	return &ListContractRequests{requests: requests}
}

func (uc *ListContractRequests) Execute(ctx context.Context, orgID string) ([]*domain.ContractRequest, error) {
	return uc.requests.ListByOrg(ctx, orgID)
}

// ListClientRequests returns a single client's own requests (portal).
type ListClientRequests struct {
	requests domain.ContractRequestRepository
}

func NewListClientRequests(requests domain.ContractRequestRepository) *ListClientRequests {
	return &ListClientRequests{requests: requests}
}

func (uc *ListClientRequests) Execute(ctx context.Context, orgID, clientID string) ([]*domain.ContractRequest, error) {
	return uc.requests.ListByClient(ctx, orgID, clientID)
}

// ApproveContractRequest sets the financial terms a manager chose and creates the
// real contract, linking it to the (now approved) request — all in one
// transaction (CreateContract's nested WithinTx reuses the outer tx).
type ApproveContractRequest struct {
	requests domain.ContractRequestRepository
	create   *CreateContract
	tx       domain.TxManager
}

func NewApproveContractRequest(requests domain.ContractRequestRepository, create *CreateContract, tx domain.TxManager) *ApproveContractRequest {
	return &ApproveContractRequest{requests: requests, create: create, tx: tx}
}

type ApproveContractRequestInput struct {
	OrgID        string
	RequestID    string
	CostPrice    money.Money
	Markup       domain.Markup
	DownPayment  money.Money
	Installments int
	Cadence      domain.Cadence
	StartDate    time.Time
}

func (uc *ApproveContractRequest) Execute(ctx context.Context, in ApproveContractRequestInput) (*domain.Contract, error) {
	var contract *domain.Contract
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		request, err := uc.requests.GetByID(ctx, in.OrgID, in.RequestID)
		if err != nil {
			return err
		}
		if request.Status() != domain.RequestPending {
			return domain.ErrRequestNotPending
		}
		contract, err = uc.create.Execute(ctx, CreateContractInput{
			OrgID:        in.OrgID,
			ClientID:     request.ClientID(),
			ProductID:    request.ProductID(),
			CostPrice:    in.CostPrice,
			Markup:       in.Markup,
			DownPayment:  in.DownPayment,
			Installments: in.Installments,
			Cadence:      in.Cadence,
			StartDate:    in.StartDate,
		})
		if err != nil {
			return err
		}
		if err := request.Approve(contract.ID()); err != nil {
			return err
		}
		return uc.requests.Save(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return contract, nil
}

// RejectContractRequest declines a pending request.
type RejectContractRequest struct {
	requests domain.ContractRequestRepository
	tx       domain.TxManager
}

func NewRejectContractRequest(requests domain.ContractRequestRepository, tx domain.TxManager) *RejectContractRequest {
	return &RejectContractRequest{requests: requests, tx: tx}
}

func (uc *RejectContractRequest) Execute(ctx context.Context, orgID, requestID string) error {
	return uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		request, err := uc.requests.GetByID(ctx, orgID, requestID)
		if err != nil {
			return err
		}
		if err := request.Reject(); err != nil {
			return err
		}
		return uc.requests.Save(ctx, request)
	})
}
