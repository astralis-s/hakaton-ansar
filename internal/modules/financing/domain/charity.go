package domain

import (
	"strings"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// CharityStatus tracks whether a sadaqa charge has been transferred to charity.
type CharityStatus string

const (
	CharityPending     CharityStatus = "pending"
	CharityTransferred CharityStatus = "transferred"
)

// CharityEntry is a fixed late-payment charge that goes to charity (sadaqa), NOT
// to the seller and NOT into the contract's Outstanding. It is a separate record
// in the charity registry — accruing it never changes the debt (anti-riba).
// Creating one is an owner-only action (enforced at the HTTP layer).
type CharityEntry struct {
	id         string
	orgID      string
	contractID string
	clientID   string
	amount     money.Money
	status     CharityStatus
	note       string
	createdBy  string
	createdAt  time.Time
}

// NewCharityEntry validates and creates a pending charity entry with a fixed
// amount (independent of how long the payment is overdue).
func NewCharityEntry(id, orgID, contractID, clientID string, amount money.Money, note, createdBy string) (CharityEntry, error) {
	if id == "" {
		return CharityEntry{}, ErrContractIDRequired
	}
	if orgID == "" {
		return CharityEntry{}, ErrOrgIDRequired
	}
	if contractID == "" {
		return CharityEntry{}, ErrContractIDRequired
	}
	if !amount.IsPositive() {
		return CharityEntry{}, ErrCharityAmountNotPositive
	}
	return CharityEntry{
		id:         id,
		orgID:      orgID,
		contractID: contractID,
		clientID:   clientID,
		amount:     amount,
		status:     CharityPending,
		note:       strings.TrimSpace(note),
		createdBy:  createdBy,
		createdAt:  time.Now().UTC(),
	}, nil
}

// RehydrateCharityEntry rebuilds a charity entry from trusted storage.
func RehydrateCharityEntry(id, orgID, contractID, clientID string, amount money.Money, status CharityStatus, note, createdBy string, createdAt time.Time) CharityEntry {
	return CharityEntry{
		id:         id,
		orgID:      orgID,
		contractID: contractID,
		clientID:   clientID,
		amount:     amount,
		status:     status,
		note:       note,
		createdBy:  createdBy,
		createdAt:  createdAt,
	}
}

func (e CharityEntry) ID() string            { return e.id }
func (e CharityEntry) OrgID() string         { return e.orgID }
func (e CharityEntry) ContractID() string    { return e.contractID }
func (e CharityEntry) ClientID() string      { return e.clientID }
func (e CharityEntry) Amount() money.Money   { return e.amount }
func (e CharityEntry) Status() CharityStatus { return e.status }
func (e CharityEntry) Note() string          { return e.note }
func (e CharityEntry) CreatedBy() string     { return e.createdBy }
func (e CharityEntry) CreatedAt() time.Time  { return e.createdAt }
