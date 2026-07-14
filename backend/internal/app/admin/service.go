// Package admin implements the platform operator's read-only backoffice —
// listados y métricas cross-tenant, solo alcanzables tras pasar por
// middleware.RequireAdminUser (Etapa post-MVP: backoffice).
package admin

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
)

var ErrCampaignNotFound = apperr.New("not_found", "No encontrado")

type Service struct {
	repo      app.AdminRepository
	campaigns app.CampaignRepository
}

func NewService(repo app.AdminRepository, campaigns app.CampaignRepository) *Service {
	return &Service{repo: repo, campaigns: campaigns}
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
		return ErrCampaignNotFound
	}
	return s.campaigns.SoftDelete(ctx, id)
}
