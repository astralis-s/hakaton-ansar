package domain

import (
	"strings"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// RequestStatus is the lifecycle of a client's contract request (заявка).
type RequestStatus string

const (
	RequestPending  RequestStatus = "pending"
	RequestApproved RequestStatus = "approved"
	RequestRejected RequestStatus = "rejected"
)

func (s RequestStatus) Valid() bool {
	return s == RequestPending || s == RequestApproved || s == RequestRejected
}

func (s RequestStatus) String() string { return string(s) }

// ContractRequest is a client's application for an installment contract: they
// pick a product and state their wishes (term, down payment, note). Staff then
// set the financial terms and, on approval, a real Contract is created and
// linked here. The request never carries riba — it is just an intent.
type ContractRequest struct {
	id                  string
	orgID               string
	clientID            string
	productID           string
	desiredInstallments int
	desiredDownPayment  money.Money
	note                string
	status              RequestStatus
	contractID          string
	createdAt           time.Time
	decidedAt           *time.Time
}

// NewContractRequest validates invariants and creates a pending request.
func NewContractRequest(id, orgID, clientID, productID string, desiredInstallments int, desiredDownPayment money.Money, note string) (*ContractRequest, error) {
	if id == "" {
		return nil, ErrRequestIDRequired
	}
	if orgID == "" {
		return nil, ErrOrgIDRequired
	}
	if clientID == "" {
		return nil, ErrClientIDRequired
	}
	if productID == "" {
		return nil, ErrProductIDRequired
	}
	if desiredInstallments < 1 {
		return nil, ErrDesiredInstallmentsInvalid
	}
	if desiredDownPayment.IsNegative() {
		return nil, ErrDownPaymentNegative
	}
	return &ContractRequest{
		id:                  id,
		orgID:               orgID,
		clientID:            clientID,
		productID:           productID,
		desiredInstallments: desiredInstallments,
		desiredDownPayment:  desiredDownPayment,
		note:                strings.TrimSpace(note),
		status:              RequestPending,
		createdAt:           time.Now().UTC(),
	}, nil
}

// RehydrateContractRequest rebuilds a request from storage.
func RehydrateContractRequest(id, orgID, clientID, productID string, desiredInstallments int, desiredDownPayment money.Money, note string, status RequestStatus, contractID string, createdAt time.Time, decidedAt *time.Time) *ContractRequest {
	return &ContractRequest{
		id:                  id,
		orgID:               orgID,
		clientID:            clientID,
		productID:           productID,
		desiredInstallments: desiredInstallments,
		desiredDownPayment:  desiredDownPayment,
		note:                note,
		status:              status,
		contractID:          contractID,
		createdAt:           createdAt,
		decidedAt:           decidedAt,
	}
}

// Approve marks a pending request approved and links the created contract.
func (r *ContractRequest) Approve(contractID string) error {
	if r.status != RequestPending {
		return ErrRequestNotPending
	}
	if contractID == "" {
		return ErrContractIDRequired
	}
	r.status = RequestApproved
	r.contractID = contractID
	now := time.Now().UTC()
	r.decidedAt = &now
	return nil
}

// Reject marks a pending request rejected.
func (r *ContractRequest) Reject() error {
	if r.status != RequestPending {
		return ErrRequestNotPending
	}
	r.status = RequestRejected
	now := time.Now().UTC()
	r.decidedAt = &now
	return nil
}

func (r *ContractRequest) ID() string                  { return r.id }
func (r *ContractRequest) OrgID() string                { return r.orgID }
func (r *ContractRequest) ClientID() string             { return r.clientID }
func (r *ContractRequest) ProductID() string            { return r.productID }
func (r *ContractRequest) DesiredInstallments() int     { return r.desiredInstallments }
func (r *ContractRequest) DesiredDownPayment() money.Money { return r.desiredDownPayment }
func (r *ContractRequest) Note() string                 { return r.note }
func (r *ContractRequest) Status() RequestStatus        { return r.status }
func (r *ContractRequest) ContractID() string           { return r.contractID }
func (r *ContractRequest) CreatedAt() time.Time         { return r.createdAt }
func (r *ContractRequest) DecidedAt() *time.Time        { return r.decidedAt }
