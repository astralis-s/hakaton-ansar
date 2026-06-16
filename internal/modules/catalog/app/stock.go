package app

import (
	"context"

	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
)

// AdjustStock changes a product's stock balance and logs a movement atomically
// (receipt, manual adjustment or write-off).
type AdjustStock struct {
	stock domain.StockRepository
	tx    domain.TxManager
}

func NewAdjustStock(stock domain.StockRepository, tx domain.TxManager) *AdjustStock {
	return &AdjustStock{stock: stock, tx: tx}
}

type AdjustStockInput struct {
	OrgID     string
	ProductID string
	Delta     int
	Reason    string
	Note      string
}

func (uc *AdjustStock) Execute(ctx context.Context, in AdjustStockInput) (domain.Product, domain.StockMovement, error) {
	reason, err := domain.ParseStockReason(in.Reason)
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, err
	}
	if in.Delta == 0 {
		return domain.Product{}, domain.StockMovement{}, domain.ErrStockDeltaZero
	}

	var product domain.Product
	var movement domain.StockMovement
	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		product, movement, err = uc.stock.Adjust(ctx, in.OrgID, in.ProductID, in.Delta, reason, in.Note)
		return err
	})
	return product, movement, err
}

// ListStockMovements returns the organization's stock movement log (товарооборот).
type ListStockMovements struct {
	stock domain.StockRepository
}

func NewListStockMovements(stock domain.StockRepository) *ListStockMovements {
	return &ListStockMovements{stock: stock}
}

func (uc *ListStockMovements) Execute(ctx context.Context, orgID string) ([]domain.StockMovement, error) {
	return uc.stock.ListMovementsByOrg(ctx, orgID)
}
