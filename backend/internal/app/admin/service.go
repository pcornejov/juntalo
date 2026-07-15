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
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

var (
	ErrNotFound          = apperr.New("not_found", "No encontrado")
	errInvalidCommission = apperr.New("invalid_commission_rate", "La comisión debe estar entre 0% y 50%")
	errInvalidPayout     = apperr.New("invalid_payout_amount", "El monto debe ser mayor a 0")
	errPayoutTooHigh     = apperr.New("payout_exceeds_pending", "El monto supera lo pendiente de liquidar para esta organización")
)

// maxCommissionRate acota lo que un admin puede cargar por el backoffice —
// un error de tipeo (ej. "50" en vez de "5") no debería poder dejar una
// organización con 5000% de comisión.
const maxCommissionRate = 0.5

type Service struct {
	repo      app.AdminRepository
	campaigns app.CampaignRepository
	orgs      app.OrganizationRepository
	payouts   app.PayoutRepository
}

func NewService(repo app.AdminRepository, campaigns app.CampaignRepository, orgs app.OrganizationRepository, payouts app.PayoutRepository) *Service {
	return &Service{repo: repo, campaigns: campaigns, orgs: orgs, payouts: payouts}
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

// ListPendingPayouts expone qué organizaciones tienen plata confirmada sin
// liquidar todavía — la vista central del backoffice para saber a quién y
// cuánto transferir manualmente.
func (s *Service) ListPendingPayouts(ctx context.Context, limit, offset int32) ([]app.PendingPayoutRow, error) {
	return s.payouts.ListPending(ctx, limit, offset)
}

// CreatePayout registra una transferencia manual ya hecha por el operador.
// No dispara ningún movimiento de dinero real — valida contra lo
// efectivamente pendiente para esa organización, para que un typo no deje
// un registro de "pagado" mayor a lo que en realidad se le debía.
func (s *Service) CreatePayout(ctx context.Context, orgID uuid.UUID, amount money.CLP, note string, createdBy uuid.UUID) (app.PayoutRecord, error) {
	if amount <= 0 {
		return app.PayoutRecord{}, errInvalidPayout
	}
	found := false
	var maxAmount money.CLP
	// ListPending no filtra por organización — se busca la fila puntual
	// entre las pendientes (a esta escala, decenas de organizaciones, es
	// más simple que sumar un endpoint dedicado por-org).
	all, err := s.payouts.ListPending(ctx, 10_000, 0)
	if err != nil {
		return app.PayoutRecord{}, err
	}
	for _, p := range all {
		if p.OrganizationID == orgID {
			found = true
			maxAmount = p.PendingAmount
			break
		}
	}
	if !found || amount > maxAmount {
		return app.PayoutRecord{}, errPayoutTooHigh
	}
	return s.payouts.Create(ctx, app.CreatePayoutInput{
		OrganizationID: orgID,
		Amount:         amount,
		Note:           note,
		CreatedBy:      createdBy,
	})
}
