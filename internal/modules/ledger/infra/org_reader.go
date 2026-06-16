package infra

import (
	"context"

	iamdomain "github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/ledger/domain"
)

// OrgReader adapts the iam organization repository to the ledger's OrgReader port
// (the organization name printed on the finance report).
type OrgReader struct {
	orgs iamdomain.OrganizationRepository
}

func NewOrgReader(orgs iamdomain.OrganizationRepository) *OrgReader {
	return &OrgReader{orgs: orgs}
}

var _ domain.OrgReader = (*OrgReader)(nil)

func (r *OrgReader) Name(ctx context.Context, orgID string) (string, error) {
	org, err := r.orgs.GetByID(ctx, orgID)
	if err != nil {
		return "", err
	}
	return org.Name(), nil
}
