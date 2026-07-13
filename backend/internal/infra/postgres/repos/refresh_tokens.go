package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type RefreshTokenRepo struct {
	q *sqlc.Queries
}

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{q: sqlc.New(pool)}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, t identity.RefreshToken) error {
	_, err := r.q.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: t.ExpiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepo) GetValidByHash(ctx context.Context, tokenHash string) (identity.RefreshToken, bool, error) {
	t, err := r.q.GetValidRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.RefreshToken{}, false, nil
		}
		return identity.RefreshToken{}, false, fmt.Errorf("get refresh token: %w", err)
	}
	return mapRefreshToken(t), true, nil
}

func (r *RefreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	if err := r.q.RevokeRefreshToken(ctx, id); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}
