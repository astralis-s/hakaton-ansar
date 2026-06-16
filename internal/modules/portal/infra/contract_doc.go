package infra

import (
	"context"
	"errors"

	financingapp "github.com/astralis-s/hakaton-ansar/internal/modules/financing/app"
	financingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/document"
)

// ContractDocBuilder adapts the financing contract-document use-case to the
// portal's ContractDocBuilder port, passing the client id so the client can only
// generate documents for their own contracts.
type ContractDocBuilder struct {
	build *financingapp.BuildContractDoc
}

func NewContractDocBuilder(build *financingapp.BuildContractDoc) *ContractDocBuilder {
	return &ContractDocBuilder{build: build}
}

var _ domain.ContractDocBuilder = (*ContractDocBuilder)(nil)

func (b *ContractDocBuilder) Build(ctx context.Context, orgID, contractID, clientID string) (document.Contract, error) {
	doc, err := b.build.Execute(ctx, orgID, contractID, clientID)
	if err != nil {
		if errors.Is(err, financingdomain.ErrContractNotFound) {
			return document.Contract{}, domain.ErrContractNotFound
		}
		return document.Contract{}, err
	}
	return doc, nil
}
