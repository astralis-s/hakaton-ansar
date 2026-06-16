package domain

import "context"

// ExpenseRepository persists manual expenses. All reads/writes are scoped to an
// organization for tenant isolation.
type ExpenseRepository interface {
	Create(ctx context.Context, e Expense) (Expense, error)
	ListByOrg(ctx context.Context, orgID string) ([]Expense, error)
	Delete(ctx context.Context, orgID, id string) error
}

// SalesReader reads the organization's sales (financing contracts) as the
// ledger's read model — just the sale/cost prices and status needed for the
// P&L. Implemented in infra over the financing context; when financing becomes
// its own service this is the only piece that changes.
type SalesReader interface {
	ListSales(ctx context.Context, orgID string) ([]Sale, error)
}

// OrgReader resolves an organization's display name (the header of the report).
type OrgReader interface {
	Name(ctx context.Context, orgID string) (string, error)
}
