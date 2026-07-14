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

type PasswordResetRepo struct {
	q *sqlc.Queries
}

func NewPasswordResetRepo(pool *pgxpool.Pool) *PasswordResetRepo {
	return &PasswordResetRepo{q: sqlc.New(pool)}
}

func (r *PasswordResetRepo) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	err := r.q.CreatePasswordResetToken(ctx, sqlc.CreatePasswordResetTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create password reset token: %w", err)
	}
	return nil
}

func (r *PasswordResetRepo) GetUserIDByValidHash(ctx context.Context, tokenHash string) (uuid.UUID, bool, error) {
	row, err := r.q.GetValidPasswordResetTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, fmt.Errorf("get password reset token: %w", err)
	}
	return row.UserID, true, nil
}

func (r *PasswordResetRepo) MarkUsed(ctx context.Context, tokenHash string) error {
	if err := r.q.MarkPasswordResetTokenUsed(ctx, tokenHash); err != nil {
		return fmt.Errorf("mark password reset token used: %w", err)
	}
	return nil
}
