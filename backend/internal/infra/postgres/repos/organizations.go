package repos

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

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
