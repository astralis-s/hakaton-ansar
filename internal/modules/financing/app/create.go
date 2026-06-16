package app

import (
	"context"
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// CreateContract validates references, builds the murabaha contract and persists
// it in Active state (create+activate atomically — the wizard creates an active
// contract; the Draft state is internal). A Haram product is refused, and one
// unit of stock is reserved in the same transaction as the contract insert.
type CreateContract struct {
	contracts domain.ContractRepository
	products  domain.ProductReader
	clients   domain.ClientReader
	stock     domain.StockReserver
	tx        domain.TxManager
}

func NewCreateContract(contracts domain.ContractRepository, products domain.ProductReader, clients domain.ClientReader, stock domain.StockReserver, tx domain.TxManager) *CreateContract {
	return &CreateContract{contracts: contracts, products: products, clients: clients, stock: stock, tx: tx}
}

type CreateContractInput struct {
	OrgID        string
	ClientID     string
	ProductID    string
	CostPrice    money.Money
	Markup       domain.Markup
	DownPayment  money.Money
	Installments int
	Cadence      domain.Cadence
	StartDate    time.Time
}

func (uc *CreateContract) Execute(ctx context.Context, in CreateContractInput) (*domain.Contract, error) {
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

	contract, err := domain.NewContract(domain.NewContractParams{
		ID:           NewID(),
		OrgID:        in.OrgID,
		ClientID:     in.ClientID,
		ProductID:    in.ProductID,
		CostPrice:    in.CostPrice,
		Markup:       in.Markup,
		DownPayment:  in.DownPayment,
		Installments: in.Installments,
		Cadence:      in.Cadence,
		StartDate:    in.StartDate,
	})
	if err != nil {
		return nil, err
	}
	if err := contract.Activate(); err != nil { // create + activate atomically
		return nil, err
	}

	if err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		// Reserve one unit first: it locks the product row and fails with
		// ErrOutOfStock before we write the contract, rolling everything back.
		if err := uc.stock.Reserve(ctx, in.OrgID, in.ProductID); err != nil {
			return err
		}
		return uc.contracts.Create(ctx, contract)
	}); err != nil {
		return nil, err
	}
	return contract, nil
}
