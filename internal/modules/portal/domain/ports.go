package domain

import (
	"context"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// AccountRepository persists client portal credentials.
type AccountRepository interface {
	// Upsert creates or replaces the account for a client (re-provisioning a
	// password just overwrites it).
	Upsert(ctx context.Context, a PortalAccount) error
	// GetByEmail looks an account up by its (globally unique) email for login.
	GetByEmail(ctx context.Context, email string) (PortalAccount, error)
	// GetByClientID returns the account for a client (to report portal access in
	// the staff UI), or ErrAccountNotFound.
	GetByClientID(ctx context.Context, orgID, clientID string) (PortalAccount, error)
}

// ChatRepository persists conversations and their messages, org-scoped.
type ChatRepository interface {
	// EnsureConversation returns the (single) conversation for a client, creating
	// it on first use.
	EnsureConversation(ctx context.Context, orgID, clientID string) (Conversation, error)
	// AppendMessage stores a message and bumps the conversation's last_message_at.
	AppendMessage(ctx context.Context, m Message) (Message, error)
	// ListMessages returns a client's thread in chronological order.
	ListMessages(ctx context.Context, orgID, clientID string) ([]Message, error)
	// ListConversations returns the org's conversations (newest activity first)
	// with a preview of the last message — the staff inbox.
	ListConversations(ctx context.Context, orgID string) ([]ConversationView, error)
}

// Hasher hashes and verifies portal passwords (bcrypt in infra).
type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

// ClientPrincipal is the authenticated client identity carried on portal requests.
type ClientPrincipal struct {
	OrgID    string
	ClientID string
}

// ClientTokenService issues and parses the client portal JWT.
type ClientTokenService interface {
	Issue(orgID, clientID string) (token string, expiresAt time.Time, err error)
	Parse(token string) (ClientPrincipal, error)
}

// ClientInfo is the minimal client data the portal needs from the crm context.
type ClientInfo struct {
	ID       string
	FullName string
	Phone    string
}

// ClientReader reads client data from the crm context (ErrClientNotFound when
// absent) and resolves display names for the staff inbox.
type ClientReader interface {
	Get(ctx context.Context, orgID, clientID string) (ClientInfo, error)
	Names(ctx context.Context, orgID string, ids []string) (map[string]string, error)
}

// ContractView is the client's read model of one of their installment contracts.
type ContractView struct {
	ID           string
	ProductID    string
	SalePrice    money.Money
	Outstanding  money.Money
	Status       string
	Installments int
	CreatedAt    time.Time
}

// InstallmentLine is one row of the client's payment schedule: how much is due,
// when, and whether it is paid / pending / overdue.
type InstallmentLine struct {
	Number  int
	DueDate time.Time
	Amount  money.Money
	Status  string // pending | partially_paid | paid | overdue
}

// PaymentLine is one payment the client has already made.
type PaymentLine struct {
	Amount money.Money
	PaidAt time.Time
}

// ContractDetail is the full client-facing view of one contract: the headline
// figures, the payment schedule (сколько и когда платить) and the payment
// history. It deliberately omits cost price / markup (the merchant's margin).
type ContractDetail struct {
	ID             string
	ProductID      string
	SalePrice      money.Money
	DownPayment    money.Money
	FinancedAmount money.Money
	Outstanding    money.Money
	PaidAmount     money.Money
	Status         string
	Cadence        string
	StartDate      time.Time
	HasOverdue     bool
	HasNext        bool
	NextDueDate    time.Time
	NextDueAmount  money.Money
	Installments   []InstallmentLine
	Payments       []PaymentLine
	CreatedAt      time.Time
}

// ContractReader reads a client's own contracts from the financing context (for
// the portal "my installments" views). GetForClient returns ErrContractNotFound
// both when the contract is absent and when it belongs to another client.
type ContractReader interface {
	ListForClient(ctx context.Context, orgID, clientID string) ([]ContractView, error)
	GetForClient(ctx context.Context, orgID, clientID, contractID string) (ContractDetail, error)
}

// ProductCard is a product as shown to the client when choosing what to request
// (no cost price — the client only picks the item; the manager quotes terms).
type ProductCard struct {
	ID       string
	Name     string
	Category string
}

// CatalogReader lists the products a client may request (halal and in stock).
type CatalogReader interface {
	ListAvailable(ctx context.Context, orgID string) ([]ProductCard, error)
}

// NewRequestInput is a client's contract application.
type NewRequestInput struct {
	OrgID               string
	ClientID            string
	ProductID           string
	DesiredInstallments int
	DesiredDownPayment  money.Money
	Note                string
}

// RequestView is the client's read model of one of their submitted requests.
type RequestView struct {
	ID                  string
	ProductID           string
	DesiredInstallments int
	DesiredDownPayment  money.Money
	Note                string
	Status              string // pending | approved | rejected
	ContractID          string
	CreatedAt           time.Time
}

// RequestService submits and lists a client's contract requests. Implemented in
// infra over the financing context (the merchant approves and sets terms there).
type RequestService interface {
	Submit(ctx context.Context, in NewRequestInput) (RequestView, error)
	ListForClient(ctx context.Context, orgID, clientID string) ([]RequestView, error)
}

// TxManager runs a function inside a single database transaction.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
