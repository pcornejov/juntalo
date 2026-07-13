package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type ContributionRepo struct {
	q *sqlc.Queries
}

func NewContributionRepo(pool *pgxpool.Pool) *ContributionRepo {
	return &ContributionRepo{q: sqlc.New(pool)}
}

func (r *ContributionRepo) Create(ctx context.Context, in app.CreateContributionInput) (contribution.Contribution, error) {
	c, err := r.q.CreateContribution(ctx, sqlc.CreateContributionParams{
		CampaignID:    in.CampaignID,
		ContributorID: in.ContributorID,
		Amount:        int64(in.Amount),
		IsAnonymous:   in.IsAnonymous,
		Message:       pgtype.Text{String: in.Message, Valid: in.Message != ""},
	})
	if err != nil {
		return contribution.Contribution{}, fmt.Errorf("create contribution: %w", err)
	}
	return mapContribution(c), nil
}

func (r *ContributionRepo) GetByID(ctx context.Context, id uuid.UUID) (contribution.Contribution, bool, error) {
	c, err := r.q.GetContributionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return contribution.Contribution{}, false, nil
		}
		return contribution.Contribution{}, false, fmt.Errorf("get contribution: %w", err)
	}
	return mapContribution(c), true, nil
}

func (r *ContributionRepo) ListByCampaign(ctx context.Context, campaignID uuid.UUID, limit, offset int32) ([]contribution.Contribution, error) {
	rows, err := r.q.ListContributionsByCampaign(ctx, sqlc.ListContributionsByCampaignParams{
		CampaignID: campaignID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list contributions: %w", err)
	}
	out := make([]contribution.Contribution, len(rows))
	for i, c := range rows {
		out[i] = mapContribution(c)
	}
	return out, nil
}
