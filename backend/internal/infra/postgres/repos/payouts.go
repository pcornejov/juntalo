package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type PayoutRepo struct {
	q *sqlc.Queries
}

func NewPayoutRepo(pool *pgxpool.Pool) *PayoutRepo {
	return &PayoutRepo{q: sqlc.New(pool)}
}

func (r *PayoutRepo) Create(ctx context.Context, in app.CreatePayoutInput) (app.PayoutRecord, error) {
	p, err := r.q.CreatePayout(ctx, sqlc.CreatePayoutParams{
		OrganizationID: in.OrganizationID,
		Amount:         int64(in.Amount),
		Note:           pgtype.Text{String: in.Note, Valid: in.Note != ""},
		CreatedBy:      in.CreatedBy,
	})
	if err != nil {
		return app.PayoutRecord{}, fmt.Errorf("create payout: %w", err)
	}
	return app.PayoutRecord{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		Amount:         money.CLP(p.Amount),
		Note:           p.Note.String,
		CreatedBy:      p.CreatedBy,
		CreatedAt:      p.CreatedAt.Time,
	}, nil
}

func (r *PayoutRepo) ListPending(ctx context.Context, limit, offset int32) ([]app.PendingPayoutRow, error) {
	rows, err := r.q.ListPendingPayouts(ctx, sqlc.ListPendingPayoutsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list pending payouts: %w", err)
	}
	out := make([]app.PendingPayoutRow, len(rows))
	for i, row := range rows {
		out[i] = app.PendingPayoutRow{
			OrganizationID:      row.OrganizationID,
			OrganizationName:    row.OrganizationName,
			Rut:                 row.Rut.String,
			PayoutBank:          row.PayoutBank.String,
			PayoutAccountType:   row.PayoutAccountType.String,
			PayoutAccountNumber: row.PayoutAccountNumber.String,
			PayoutHolderName:    row.PayoutHolderName.String,
			EligibleNet:         money.CLP(row.EligibleNet),
			TotalPaid:           money.CLP(row.TotalPaid),
			PendingAmount:       money.CLP(row.PendingAmount),
		}
	}
	return out, nil
}
