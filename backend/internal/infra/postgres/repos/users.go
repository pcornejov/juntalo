package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type UserRepo struct {
	q *sqlc.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{q: sqlc.New(pool)}
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (identity.User, bool, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.User{}, false, nil
		}
		return identity.User{}, false, fmt.Errorf("get user by email: %w", err)
	}
	return mapUser(u), true, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (identity.User, bool, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.User{}, false, nil
		}
		return identity.User{}, false, fmt.Errorf("get user by id: %w", err)
	}
	return mapUser(u), true, nil
}
