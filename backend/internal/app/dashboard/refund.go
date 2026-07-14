package dashboard

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

var ErrContributionNotFound = apperr.New("contribution_not_found", "Aporte no encontrado")

type RefundService struct {
	campaigns     app.CampaignRepository
	contributions app.ContributionRepository
	payments      app.PaymentRepository
	provider      app.PaymentProvider
}

func NewRefundService(campaigns app.CampaignRepository, contributions app.ContributionRepository, payments app.PaymentRepository, provider app.PaymentProvider) *RefundService {
	return &RefundService{campaigns: campaigns, contributions: contributions, payments: payments, provider: provider}
}

// Refund scopes to the organizer's own campaign — igual que ParticipantsService,
// un contributionID ajeno o inexistente responde campaign_not_found /
// contribution_not_found, nunca un error de permisos (Etapa 4 §3).
func (s *RefundService) Refund(ctx context.Context, campaignID, contributionID, orgID uuid.UUID, amount money.CLP, reason string) (payment.Payment, error) {
	if _, found, err := s.campaigns.GetByIDForOrg(ctx, campaignID, orgID); err != nil {
		return payment.Payment{}, err
	} else if !found {
		return payment.Payment{}, ErrCampaignNotFound
	}

	contrib, found, err := s.contributions.GetByID(ctx, contributionID)
	if err != nil {
		return payment.Payment{}, err
	}
	if !found || contrib.CampaignID != campaignID {
		return payment.Payment{}, ErrContributionNotFound
	}

	pay, found, err := s.payments.GetByContributionID(ctx, contributionID)
	if err != nil {
		return payment.Payment{}, err
	}
	if !found {
		return payment.Payment{}, apperr.New("payment_not_found", "Pago no encontrado")
	}

	// El reembolso real en la pasarela va PRIMERO y fuera de cualquier
	// transacción de DB — es una llamada de red que no debería sostener un
	// FOR UPDATE, y no tiene sentido marcar "reembolsado" en Juntalo si el
	// dinero nunca volvió de verdad al medio de pago del aportante.
	result, err := s.provider.Refund(ctx, pay.ProviderRef, amount)
	if err != nil {
		return payment.Payment{}, apperr.New("refund_provider_failed", "No se pudo procesar el reembolso con la pasarela de pago")
	}

	return s.payments.Refund(ctx, pay.ID, amount, result.ProviderRef, reason)
}
