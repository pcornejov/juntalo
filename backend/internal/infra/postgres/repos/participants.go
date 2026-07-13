package repos

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type ParticipantRepo struct {
	q *sqlc.Queries
}

func NewParticipantRepo(pool *pgxpool.Pool) *ParticipantRepo {
	return &ParticipantRepo{q: sqlc.New(pool)}
}

func (r *ParticipantRepo) ListByCampaign(ctx context.Context, campaignID uuid.UUID, limit, offset int32) ([]app.ParticipantRow, error) {
	rows, err := r.q.ListParticipantsByCampaign(ctx, sqlc.ListParticipantsByCampaignParams{
		CampaignID: campaignID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}

	out := make([]app.ParticipantRow, len(rows))
	for i, row := range rows {
		out[i] = app.ParticipantRow{
			ContributionID: row.ContributionID,
			FullName:       row.ContributorFullName,
			Email:          row.ContributorEmail.String,
			Phone:          row.ContributorPhone.String,
			Amount:         money.CLP(row.Amount),
			RefundedAmount: money.CLP(row.RefundedAmount),
			IsAnonymous:    row.IsAnonymous,
			Status:         contribution.Status(row.Status),
			CreatedAt:      row.CreatedAt.Time,
		}
	}
	return out, nil
}
