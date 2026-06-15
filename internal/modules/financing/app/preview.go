package app

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
)

// PreviewContract computes a murabaha schedule without creating or persisting an
// aggregate (used by the contract wizard: POST /api/app/contracts/preview).
type PreviewContract struct {
	comparisonAnnualRatePercent decimal.Decimal
}

func NewPreviewContract(comparisonAnnualRatePercent decimal.Decimal) *PreviewContract {
	return &PreviewContract{comparisonAnnualRatePercent: comparisonAnnualRatePercent}
}

func (uc *PreviewContract) Execute(_ context.Context, in domain.PreviewInput) (domain.PreviewResult, error) {
	return domain.Preview(in, uc.comparisonAnnualRatePercent)
}
