package repos

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type ContributorRepo struct {
	q *sqlc.Queries
}

func NewContributorRepo(pool *pgxpool.Pool) *ContributorRepo {
	return &ContributorRepo{q: sqlc.New(pool)}
}

func (r *ContributorRepo) Create(ctx context.Context, in app.CreateContributorInput) (uuid.UUID, error) {
	c, err := r.q.CreateContributor(ctx, sqlc.CreateContributorParams{
		FullName: in.FullName,
		Email:    pgtype.Text{String: in.Email, Valid: in.Email != ""},
		Phone:    pgtype.Text{String: in.Phone, Valid: in.Phone != ""},
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create contributor: %w", err)
	}
	return c.ID, nil
}
