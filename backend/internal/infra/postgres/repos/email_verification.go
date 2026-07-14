package repos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type EmailVerificationRepo struct {
	q *sqlc.Queries
}

func NewEmailVerificationRepo(pool *pgxpool.Pool) *EmailVerificationRepo {
	return &EmailVerificationRepo{q: sqlc.New(pool)}
}

func (r *EmailVerificationRepo) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	err := r.q.CreateEmailVerificationToken(ctx, sqlc.CreateEmailVerificationTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create email verification token: %w", err)
	}
	return nil
}

func (r *EmailVerificationRepo) GetUserIDByValidHash(ctx context.Context, tokenHash string) (uuid.UUID, bool, error) {
	row, err := r.q.GetValidEmailVerificationTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, fmt.Errorf("get email verification token: %w", err)
	}
	return row.UserID, true, nil
}

func (r *EmailVerificationRepo) MarkUsed(ctx context.Context, tokenHash string) error {
	if err := r.q.MarkEmailVerificationTokenUsed(ctx, tokenHash); err != nil {
		return fmt.Errorf("mark email verification token used: %w", err)
	}
	return nil
}
