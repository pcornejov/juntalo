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
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type OrganizationRepo struct {
	q *sqlc.Queries
}

func NewOrganizationRepo(pool *pgxpool.Pool) *OrganizationRepo {
	return &OrganizationRepo{q: sqlc.New(pool)}
}

func (r *OrganizationRepo) GetPersonalByUserID(ctx context.Context, userID uuid.UUID) (identity.Organization, error) {
	o, err := r.q.GetPersonalOrganizationByUserID(ctx, userID)
	if err != nil {
		return identity.Organization{}, fmt.Errorf("get personal organization: %w", err)
	}
	return mapOrganization(o), nil
}

func (r *OrganizationRepo) GetByID(ctx context.Context, id uuid.UUID) (identity.Organization, error) {
	o, err := r.q.GetOrganizationByID(ctx, id)
	if err != nil {
		return identity.Organization{}, fmt.Errorf("get organization: %w", err)
	}
	return mapOrganization(o), nil
}

func (r *OrganizationRepo) GetBySlug(ctx context.Context, slug string) (identity.Organization, bool, error) {
	o, err := r.q.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.Organization{}, false, nil
		}
		return identity.Organization{}, false, fmt.Errorf("get organization by slug: %w", err)
	}
	return mapOrganization(o), true, nil
}

// GetOwnerEmail busca el dueño ('owner') de la organización — a esta escala
// (sin equipos/UI de invitaciones todavía) cada organización tiene
// exactamente un miembro, pero igual filtramos por rol para no depender de
// ese supuesto implícito.
func (r *OrganizationRepo) GetOwnerEmail(ctx context.Context, id uuid.UUID) (email, fullName string, err error) {
	row, err := r.q.GetOrganizationOwnerByOrgID(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("get organization owner: %w", err)
	}
	return row.Email, row.FullName, nil
}

// GetOwnerInfo agrega IsVerified (email del dueño verificado) sobre
// GetOwnerEmail para el badge de verificación de la página pública.
func (r *OrganizationRepo) GetOwnerInfo(ctx context.Context, id uuid.UUID) (fullName string, isVerified bool, err error) {
	row, err := r.q.GetOrganizationOwnerByOrgID(ctx, id)
	if err != nil {
		return "", false, fmt.Errorf("get organization owner: %w", err)
	}
	return row.FullName, row.EmailVerifiedAt.Valid, nil
}

func (r *OrganizationRepo) UpdatePayoutInfo(ctx context.Context, id uuid.UUID, in app.UpdatePayoutInfoInput) (identity.Organization, error) {
	o, err := r.q.UpdateOrganizationPayoutInfo(ctx, sqlc.UpdateOrganizationPayoutInfoParams{
		ID:                  id,
		Rut:                 pgtype.Text{String: in.Rut, Valid: in.Rut != ""},
		PayoutBank:          pgtype.Text{String: in.PayoutBank, Valid: in.PayoutBank != ""},
		PayoutAccountType:   pgtype.Text{String: in.PayoutAccountType, Valid: in.PayoutAccountType != ""},
		PayoutAccountNumber: pgtype.Text{String: in.PayoutAccountNumber, Valid: in.PayoutAccountNumber != ""},
		PayoutHolderName:    pgtype.Text{String: in.PayoutHolderName, Valid: in.PayoutHolderName != ""},
	})
	if err != nil {
		return identity.Organization{}, fmt.Errorf("update organization payout info: %w", err)
	}
	return mapOrganization(o), nil
}

func (r *OrganizationRepo) UpdateCommissionRate(ctx context.Context, id uuid.UUID, rate float64) (identity.Organization, bool, error) {
	o, err := r.q.UpdateOrganizationCommissionRate(ctx, sqlc.UpdateOrganizationCommissionRateParams{
		ID:             id,
		CommissionRate: toNumeric(rate),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.Organization{}, false, nil
		}
		return identity.Organization{}, false, fmt.Errorf("update organization commission rate: %w", err)
	}
	return mapOrganization(o), true, nil
}
