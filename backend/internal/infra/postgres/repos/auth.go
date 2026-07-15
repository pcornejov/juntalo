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
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

// orgSlugSuffixLen mirrors campaign.WithSuffix's 4-char UUID suffix, but
// generamos el slug de organización siempre con sufijo (no lo intentamos
// primero sin él): el nombre por defecto "Organización de <nombre>" repite
// mucho entre usuarios distintos, así que la colisión sin sufijo sería
// frecuente en vez de excepcional.
func newOrgSlug(name string) string {
	base := identity.SlugifyOrgName(name)
	return fmt.Sprintf("%s-%s", base, uuid.New().String()[:6])
}

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

		orgName := "Organización de " + in.FullName
		o, err := q.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
			Name: orgName,
			Slug: newOrgSlug(orgName),
		})
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

func (r *AuthRepo) GetUserIDByGoogleSubject(ctx context.Context, subject string) (uuid.UUID, bool, error) {
	row, err := sqlc.New(r.pool).GetUserIdentityByGoogleSubject(ctx, pgtype.Text{String: subject, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, fmt.Errorf("get user identity by google subject: %w", err)
	}
	return row.UserID, true, nil
}

func (r *AuthRepo) LinkGoogleIdentity(ctx context.Context, userID uuid.UUID, subject string) error {
	_, err := sqlc.New(r.pool).CreateGoogleUserIdentity(ctx, sqlc.CreateGoogleUserIdentityParams{
		UserID:          userID,
		ProviderSubject: pgtype.Text{String: subject, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("link google identity: %w", err)
	}
	return nil
}

func (r *AuthRepo) RegisterGoogle(ctx context.Context, in app.RegisterGoogleInput) (identity.User, identity.Organization, error) {
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

		_, err = q.CreateGoogleUserIdentity(ctx, sqlc.CreateGoogleUserIdentityParams{
			UserID:          u.ID,
			ProviderSubject: pgtype.Text{String: in.Subject, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("create google identity: %w", err)
		}

		orgName := "Organización de " + in.FullName
		o, err := q.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
			Name: orgName,
			Slug: newOrgSlug(orgName),
		})
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
