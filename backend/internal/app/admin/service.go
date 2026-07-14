// Package admin implements the platform operator's read-only backoffice —
// listados y métricas cross-tenant, solo alcanzables tras pasar por
// middleware.RequireAdminUser (Etapa post-MVP: backoffice).
package admin

import (
	"context"

	"github.com/pcornejov/juntalo/backend/internal/app"
)

type Service struct {
	repo app.AdminRepository
}

func NewService(repo app.AdminRepository) *Service {
	return &Service{repo: repo}
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
