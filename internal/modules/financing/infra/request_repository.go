package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing/infra/sqlcgen"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/pgconv"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// RequestRepository implements domain.ContractRequestRepository over sqlc.
type RequestRepository struct{ pool *pgxpool.Pool }

func NewRequestRepository(pool *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{pool: pool}
}

var _ domain.ContractRequestRepository = (*RequestRepository)(nil)

func (r *RequestRepository) q(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(database.Querier(ctx, r.pool))
}

func (r *RequestRepository) Create(ctx context.Context, req *domain.ContractRequest) error {
	id, err := pgconv.UUID(req.ID())
	if err != nil {
		return fmt.Errorf("invalid request id: %w", err)
	}
	org, err := pgconv.UUID(req.OrgID())
	if err != nil {
		return fmt.Errorf("invalid org id: %w", err)
	}
	cid, err := pgconv.UUID(req.ClientID())
	if err != nil {
		return fmt.Errorf("invalid client id: %w", err)
	}
	pid, err := pgconv.UUID(req.ProductID())
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}
	down, err := pgconv.Numeric(req.DesiredDownPayment().Amount())
	if err != nil {
		return err
	}
	if _, err := r.q(ctx).CreateContractRequest(ctx, sqlcgen.CreateContractRequestParams{
		ID:                  id,
		OrgID:               org,
		ClientID:            cid,
		ProductID:           pid,
		DesiredInstallments: int32(req.DesiredInstallments()),
		DesiredDownPayment:  down,
		Note:                req.Note(),
		Status:              req.Status().String(),
	}); err != nil {
		return fmt.Errorf("create contract request: %w", err)
	}
	return nil
}

func (r *RequestRepository) GetByID(ctx context.Context, orgID, id string) (*domain.ContractRequest, error) {
	rid, err := pgconv.UUID(id)
	if err != nil {
		return nil, domain.ErrRequestNotFound
	}
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, domain.ErrRequestNotFound
	}
	row, err := r.q(ctx).GetContractRequestByID(ctx, sqlcgen.GetContractRequestByIDParams{ID: rid, OrgID: org})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRequestNotFound
		}
		return nil, fmt.Errorf("get contract request: %w", err)
	}
	return requestFromRow(row)
}

func (r *RequestRepository) ListByOrg(ctx context.Context, orgID string) ([]*domain.ContractRequest, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	rows, err := r.q(ctx).ListContractRequestsByOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("list contract requests: %w", err)
	}
	return requestsFromRows(rows)
}

func (r *RequestRepository) ListByClient(ctx context.Context, orgID, clientID string) ([]*domain.ContractRequest, error) {
	org, err := pgconv.UUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid org id: %w", err)
	}
	cid, err := pgconv.UUID(clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client id: %w", err)
	}
	rows, err := r.q(ctx).ListContractRequestsByClient(ctx, sqlcgen.ListContractRequestsByClientParams{OrgID: org, ClientID: cid})
	if err != nil {
		return nil, fmt.Errorf("list client contract requests: %w", err)
	}
	return requestsFromRows(rows)
}

func (r *RequestRepository) Save(ctx context.Context, req *domain.ContractRequest) error {
	rid, err := pgconv.UUID(req.ID())
	if err != nil {
		return domain.ErrRequestNotFound
	}
	org, err := pgconv.UUID(req.OrgID())
	if err != nil {
		return domain.ErrRequestNotFound
	}
	contractID, err := pgconv.NullableUUID(req.ContractID())
	if err != nil {
		return fmt.Errorf("invalid contract id: %w", err)
	}
	if err := r.q(ctx).UpdateContractRequest(ctx, sqlcgen.UpdateContractRequestParams{
		ID:         rid,
		OrgID:      org,
		Status:     req.Status().String(),
		ContractID: contractID,
		DecidedAt:  pgconv.NullableTimestamp(req.DecidedAt()),
	}); err != nil {
		return fmt.Errorf("update contract request: %w", err)
	}
	return nil
}

func requestsFromRows(rows []sqlcgen.ContractRequest) ([]*domain.ContractRequest, error) {
	out := make([]*domain.ContractRequest, 0, len(rows))
	for _, row := range rows {
		req, err := requestFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, req)
	}
	return out, nil
}

func requestFromRow(row sqlcgen.ContractRequest) (*domain.ContractRequest, error) {
	down, err := pgconv.DecimalFromNumeric(row.DesiredDownPayment)
	if err != nil {
		return nil, fmt.Errorf("decode desired down payment: %w", err)
	}
	return domain.RehydrateContractRequest(
		pgconv.StrUUID(row.ID),
		pgconv.StrUUID(row.OrgID),
		pgconv.StrUUID(row.ClientID),
		pgconv.StrUUID(row.ProductID),
		int(row.DesiredInstallments),
		money.New(down, money.DefaultCurrency),
		row.Note,
		domain.RequestStatus(row.Status),
		pgconv.StrUUID(row.ContractID),
		pgconv.TimeValue(row.CreatedAt),
		pgconv.TimePtr(row.DecidedAt),
	), nil
}
