// Package admin implements el backoffice del operador de la plataforma —
// listados/métricas cross-tenant y herramientas de moderación (eliminar
// campañas, ajustar comisión), solo alcanzables tras pasar por
// middleware.RequireAdminUser (Etapa post-MVP: backoffice).
package admin

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

var (
	ErrNotFound          = apperr.New("not_found", "No encontrado")
	errInvalidCommission = apperr.New("invalid_commission_rate", "La comisión debe estar entre 0% y 50%")
)

// maxCommissionRate acota lo que un admin puede cargar por el backoffice —
// un error de tipeo (ej. "50" en vez de "5") no debería poder dejar una
// organización con 5000% de comisión.
const maxCommissionRate = 0.5

type Service struct {
	repo      app.AdminRepository
	campaigns app.CampaignRepository
	orgs      app.OrganizationRepository
}

func NewService(repo app.AdminRepository, campaigns app.CampaignRepository, orgs app.OrganizationRepository) *Service {
	return &Service{repo: repo, campaigns: campaigns, orgs: orgs}
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int32) ([]app.AdminUserRow, error) {
	return s.repo.ListUsers(ctx, limit, offset)
}

func (s *Service) ListCampaigns(ctx context.Context, limit, offset int32) ([]app.AdminCampaignRow, error) {
	return s.repo.ListCampaigns(ctx, limit, offset)
}

func (s *Service) ListPayments(ctx context.Context, limit, offset int32) ([]app.AdminPaymentRow, error) {
	return s.repo.ListPayments(ctx, limit, offset)
}

func (s *Service) Metrics(ctx context.Context) (app.AdminMetrics, error) {
	return s.repo.GetMetrics(ctx)
}

// DeleteCampaign es la herramienta de moderación del backoffice: a
// diferencia de campaigns.DeleteService (que solo permite borrar borradores
// propios), un admin puede eliminar cualquier campaña de cualquier
// organización sin importar su estado — misma respuesta 404 si el id no
// existe, nunca un error de permisos.
func (s *Service) DeleteCampaign(ctx context.Context, id uuid.UUID) error {
	_, found, err := s.campaigns.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	return s.campaigns.SoftDelete(ctx, id)
}

// UpdateOrgCommissionRate deja ajustar la comisión de una organización
// puntual desde el backoffice, sin tocar código — rate como fracción
// (0.05 = 5%), igual que el resto del dominio de payment.
func (s *Service) UpdateOrgCommissionRate(ctx context.Context, orgID uuid.UUID, rate float64) (identity.Organization, error) {
	if rate < 0 || rate > maxCommissionRate {
		return identity.Organization{}, errInvalidCommission
	}
	org, found, err := s.orgs.UpdateCommissionRate(ctx, orgID, rate)
	if err != nil {
		return identity.Organization{}, err
	}
	if !found {
		return identity.Organization{}, ErrNotFound
	}
	return org, nil
}
