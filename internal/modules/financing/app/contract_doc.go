package app

import (
	"context"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/document"
)

// BuildContractDoc gathers the denormalized data for a contract agreement
// document (contract figures + client/product/seller names + schedule).
type BuildContractDoc struct {
	contracts domain.ContractRepository
	products  domain.ProductReader
	clients   domain.ClientReader
	orgs      domain.OrgReader
}

func NewBuildContractDoc(contracts domain.ContractRepository, products domain.ProductReader, clients domain.ClientReader, orgs domain.OrgReader) *BuildContractDoc {
	return &BuildContractDoc{contracts: contracts, products: products, clients: clients, orgs: orgs}
}

// Execute returns the document data. When requireClientID is non-empty the
// contract must belong to that client (the portal passes the logged-in client;
// staff pass ""), otherwise ErrContractNotFound is returned.
func (uc *BuildContractDoc) Execute(ctx context.Context, orgID, contractID, requireClientID string) (document.Contract, error) {
	c, err := uc.contracts.GetByID(ctx, orgID, contractID)
	if err != nil {
		return document.Contract{}, err
	}
	if requireClientID != "" && c.ClientID() != requireClientID {
		return document.Contract{}, domain.ErrContractNotFound
	}

	doc := document.Contract{
		Number:            shortID(c.ID()),
		Date:              c.CreatedAt().Format(docDate),
		Status:            contractStatusLabel(c.Status()),
		CostPrice:         c.CostPrice().String(),
		Markup:            c.Markup().Money().String(),
		SalePrice:         c.SalePrice().String(),
		DownPayment:       c.DownPayment().String(),
		FinancedAmount:    c.FinancedAmount().String(),
		Outstanding:       c.Outstanding().String(),
		PaidAmount:        c.PaidAmount().String(),
		InstallmentsCount: c.InstallmentsCount(),
		Cadence:           cadenceLabel(c.Cadence()),
		StartDate:         c.StartDate().Format(docDate),
	}

	if name, err := uc.orgs.Name(ctx, orgID); err == nil {
		doc.OrgName = name
	}
	if pi, err := uc.products.Get(ctx, orgID, c.ProductID()); err == nil {
		doc.ProductName = pi.Name
	}
	if ct, err := uc.clients.Contact(ctx, orgID, c.ClientID()); err == nil {
		doc.ClientName = ct.Name
		doc.ClientPhone = ct.Phone
		doc.ClientDocument = ct.Document
	}

	asOf := time.Now()
	for _, v := range c.Installments(asOf) {
		doc.Schedule = append(doc.Schedule, document.ContractLine{
			Number:  v.Number,
			DueDate: v.DueDate.Format(docDate),
			Amount:  v.Amount.String(),
			Status:  installmentStatusLabel(v.Status),
		})
	}
	return doc, nil
}

const docDate = "02.01.2006"

func shortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

func cadenceLabel(c domain.Cadence) string {
	if c == domain.CadenceWeekly {
		return "еженедельно"
	}
	return "ежемесячно"
}

func contractStatusLabel(s domain.ContractStatus) string {
	switch s {
	case domain.StatusActive:
		return "Активен"
	case domain.StatusCompleted:
		return "Завершён"
	case domain.StatusCancelled:
		return "Отменён"
	case domain.StatusDraft:
		return "Черновик"
	}
	return string(s)
}

func installmentStatusLabel(s domain.InstallmentStatus) string {
	switch s {
	case domain.InstallmentPaid:
		return "Оплачен"
	case domain.InstallmentPartiallyPaid:
		return "Частично"
	case domain.InstallmentOverdue:
		return "Просрочен"
	case domain.InstallmentPending:
		return "Предстоит"
	}
	return string(s)
}
