package repos

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

// AuthRepo implements app.AuthRepository. Register runs the user+identity+
// organization+membership transaction described in Etapa 3 §2.
type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) Register(ctx context.Context, in app.RegisterInput) (identity.User, identity.Organization, error) {
	var user identity.User
	var org identity.Organization

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		q := sqlc.New(tx)

		u, err := q.CreateUser(ctx, sqlc.CreateUserParams{Email: in.Email, FullName: in.FullName})
		if err != nil {
			if isUniqueViolation(err) {
				return apperr.New("email_already_registered", "Ya existe una cuenta con este email")
			}
			return fmt.Errorf("create user: %w", err)
		}

		_, err = q.CreateUserIdentity(ctx, sqlc.CreateUserIdentityParams{
			UserID:       u.ID,
			PasswordHash: pgtype.Text{String: in.PasswordHash, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("create identity: %w", err)
		}

		o, err := q.CreateOrganization(ctx, "Organización de "+in.FullName)
		if err != nil {
			return fmt.Errorf("create organization: %w", err)
		}

		err = q.CreateOrganizationMember(ctx, sqlc.CreateOrganizationMemberParams{
			OrganizationID: o.ID,
			UserID:         u.ID,
			Role:           "owner",
		})
		if err != nil {
			return fmt.Errorf("create membership: %w", err)
		}

		user = mapUser(u)
		org = mapOrganization(o)
		return nil
	})
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}
	return user, org, nil
}

func (r *AuthRepo) GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	identityRow, err := sqlc.New(r.pool).GetPasswordIdentityByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get password identity: %w", err)
	}
	return identityRow.PasswordHash.String, nil
}

func (r *AuthRepo) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string) error {
	err := sqlc.New(r.pool).UpdatePasswordHash(ctx, sqlc.UpdatePasswordHashParams{
		UserID:       userID,
		PasswordHash: pgtype.Text{String: newHash, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	return nil
}
