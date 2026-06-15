// Package infra provides the financing persistence adapters (sqlc repositories)
// and the cross-context reader adapters (catalog product, crm client).
package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// ContractRepository implements domain.ContractRepository over sqlc.
type ContractRepository struct{ pool *pgxpool.Pool }

func NewContractRepository(pool *pgxpool.Pool) *ContractRepository {
	return &ContractRepository{pool: pool}
}

var _ domain.ContractRepository = (*ContractRepository)(nil)

func (r *ContractRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *ContractRepository) Create(ctx context.Context, c *domain.Contract) error {
	id, err := pgconv.UUID(c.ID())
	if err != nil {
		return fmt.Errorf("invalid contract id: %w", err)
	}
	orgID, err := pgconv.UUID(c.OrgID())
	if err != nil {
		return fmt.Errorf("invalid org id: %w", err)
	}
	clientID, err := pgconv.UUID(c.ClientID())
	if err != nil {
		return fmt.Errorf("invalid client id: %w", err)
	}
	productID, err := pgconv.UUID(c.ProductID())
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}
	nums, err := contractNumerics(c)
	if err != nil {
		return err
	}

	if err := r.q(ctx).CreateContract(ctx, sqlcgen.CreateContractParams{
		ID:                id,
		OrgID:             orgID,
		ClientID:          clientID,
		ProductID:         productID,
		CostPrice:         nums.cost,
		Markup:            nums.markup,
		SalePrice:         nums.sale,
		DownPayment:       nums.down,
		FinancedAmount:    nums.financed,
		Outstanding:       nums.outstanding,
		InstallmentsCount: int32(c.InstallmentsCount()),
		Cadence:           c.Cadence().String(),
		Currency:          c.Outstanding().Currency(),
		Status:            c.Status().String(),
		StartDate:         pgconv.Date(c.StartDate()),
	}); err != nil {
		return fmt.Errorf("insert contract: %w", err)
	}

	for _, inst := range c.Schedule() {
		instID, err := pgconv.UUID(newID())
		if err != nil {
			return err
		}
		amount, err := pgconv.Numeric(inst.Amount().Amount())
		if err != nil {
			return err
		}
		if err := r.q(ctx).CreateInstallment(ctx, sqlcgen.CreateInstallmentParams{
			ID:         instID,
			ContractID: id,
			Number:     int32(inst.Number()),
			DueDate:    pgconv.Date(inst.DueDate()),
			Amount:     amount,
		}); err != nil {
			return fmt.Errorf("insert installment: %w", err)
		}
	}
	return nil
}

func (r *ContractRepository) GetByID(ctx context.Context, orgID, id string) (*domain.Contract, error) {
	cid, err := pgconv.UUID(id)
	if err != nil {
		return nil, domain.ErrContractNotFound
	}
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, domain.ErrContractNotFound
	}

	row, err := r.q(ctx).GetContractByID(ctx, sqlcgen.GetContractByIDParams{ID: cid, OrgID: org})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContractNotFound
		}
		return nil, fmt.Errorf("get contract: %w", err)
	}

	instRows, err := r.q(ctx).ListInstallmentsByContract(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("list installments: %w", err)
	}
	payRows, err := r.q(ctx).ListPaymentsByContract(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	return contractFromRows(row, instRows, payRows)
}

func (r *ContractRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.ContractSummary, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.q(ctx).ListContractsByOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	summaries := make([]domain.ContractSummary, 0, len(rows))
	for _, row := range rows {
		sale, err := moneyFrom(row.SalePrice, row.Currency)
		if err != nil {
			return nil, err
		}
		financed, err := moneyFrom(row.FinancedAmount, row.Currency)
		if err != nil {
			return nil, err
		}
		outstanding, err := moneyFrom(row.Outstanding, row.Currency)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, domain.ContractSummary{
			ID:                pgconv.StrUUID(row.ID),
			ClientID:          pgconv.StrUUID(row.ClientID),
			ProductID:         pgconv.StrUUID(row.ProductID),
			SalePrice:         sale,
			FinancedAmount:    financed,
			Outstanding:       outstanding,
			Status:            domain.ContractStatus(row.Status),
			InstallmentsCount: int(row.InstallmentsCount),
			CreatedAt:         pgconv.TimeValue(row.CreatedAt),
		})
	}
	return summaries, nil
}

func (r *ContractRepository) SaveState(ctx context.Context, c *domain.Contract) error {
	id, err := pgconv.UUID(c.ID())
	if err != nil {
		return fmt.Errorf("invalid contract id: %w", err)
	}
	org, err := pgconv.UUID(c.OrgID())
	if err != nil {
		return fmt.Errorf("invalid org id: %w", err)
	}
	outstanding, err := pgconv.Numeric(c.Outstanding().Amount())
	if err != nil {
		return err
	}
	if err := r.q(ctx).UpdateContractState(ctx, sqlcgen.UpdateContractStateParams{
		ID:          id,
		OrgID:       org,
		Outstanding: outstanding,
		Status:      c.Status().String(),
	}); err != nil {
		return fmt.Errorf("update contract state: %w", err)
	}
	return nil
}

func (r *ContractRepository) AddPayment(ctx context.Context, contractID string, p domain.Payment) error {
	pid, err := pgconv.UUID(p.ID())
	if err != nil {
		return fmt.Errorf("invalid payment id: %w", err)
	}
	cid, err := pgconv.UUID(contractID)
	if err != nil {
		return fmt.Errorf("invalid contract id: %w", err)
	}
	amount, err := pgconv.Numeric(p.Amount().Amount())
	if err != nil {
		return err
	}
	if err := r.q(ctx).CreatePayment(ctx, sqlcgen.CreatePaymentParams{
		ID:         pid,
		ContractID: cid,
		Amount:     amount,
		PaidAt:     pgconv.Timestamp(p.PaidAt()),
	}); err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

// --- row → domain mapping ---------------------------------------------------

type contractNums struct {
	cost, markup, sale, down, financed, outstanding pgtype.Numeric
}

func contractNumerics(c *domain.Contract) (contractNums, error) {
	var n contractNums
	var err error
	if n.cost, err = pgconv.Numeric(c.CostPrice().Amount()); err != nil {
		return n, err
	}
	if n.markup, err = pgconv.Numeric(c.Markup().Money().Amount()); err != nil {
		return n, err
	}
	if n.sale, err = pgconv.Numeric(c.SalePrice().Amount()); err != nil {
		return n, err
	}
	if n.down, err = pgconv.Numeric(c.DownPayment().Amount()); err != nil {
		return n, err
	}
	if n.financed, err = pgconv.Numeric(c.FinancedAmount().Amount()); err != nil {
		return n, err
	}
	if n.outstanding, err = pgconv.Numeric(c.Outstanding().Amount()); err != nil {
		return n, err
	}
	return n, nil
}

func moneyFrom(n pgtype.Numeric, currency string) (money.Money, error) {
	d, err := pgconv.DecimalFromNumeric(n)
	if err != nil {
		return money.Money{}, err
	}
	return money.New(d, currency), nil
}

func contractFromRows(row sqlcgen.Contract, instRows []sqlcgen.Installment, payRows []sqlcgen.Payment) (*domain.Contract, error) {
	currency := row.Currency

	cost, err := moneyFrom(row.CostPrice, currency)
	if err != nil {
		return nil, err
	}
	markupMoney, err := moneyFrom(row.Markup, currency)
	if err != nil {
		return nil, err
	}
	markup, err := domain.NewMarkup(markupMoney)
	if err != nil {
		return nil, err
	}
	sale, err := moneyFrom(row.SalePrice, currency)
	if err != nil {
		return nil, err
	}
	down, err := moneyFrom(row.DownPayment, currency)
	if err != nil {
		return nil, err
	}
	financed, err := moneyFrom(row.FinancedAmount, currency)
	if err != nil {
		return nil, err
	}
	outstanding, err := moneyFrom(row.Outstanding, currency)
	if err != nil {
		return nil, err
	}

	schedule := make([]domain.Installment, 0, len(instRows))
	for _, ir := range instRows {
		amount, err := moneyFrom(ir.Amount, currency)
		if err != nil {
			return nil, err
		}
		schedule = append(schedule, domain.NewInstallment(int(ir.Number), pgconv.DateValue(ir.DueDate), amount))
	}

	payments := make([]domain.Payment, 0, len(payRows))
	for _, pr := range payRows {
		amount, err := moneyFrom(pr.Amount, currency)
		if err != nil {
			return nil, err
		}
		payments = append(payments, domain.NewPayment(pgconv.StrUUID(pr.ID), amount, pgconv.TimeValue(pr.PaidAt)))
	}

	return domain.RehydrateContract(
		pgconv.StrUUID(row.ID),
		pgconv.StrUUID(row.OrgID),
		pgconv.StrUUID(row.ClientID),
		pgconv.StrUUID(row.ProductID),
		cost, markup, sale, down, financed, outstanding,
		schedule, payments,
		domain.ContractStatus(row.Status),
		domain.Cadence(row.Cadence),
		pgconv.DateValue(row.StartDate),
		pgconv.TimeValue(row.CreatedAt),
	), nil
}
