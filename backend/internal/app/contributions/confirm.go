package contributions

import (
	"context"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

// IsPaymentNotFound reports whether err is the apperr raised when a webhook
// references a provider_ref that doesn't exist — treated as harmless noise
// by the handler (Etapa 4 §5), not a real failure.
func IsPaymentNotFound(err error) bool {
	return apperr.Is(err, "payment_not_found")
}

type ConfirmService struct {
	payments      app.PaymentRepository
	contributions app.ContributionRepository
	campaigns     app.CampaignRepository
	contributors  app.ContributorRepository
	organizations app.OrganizationRepository
	email         app.EmailSender
	frontendURL   string
}

func NewConfirmService(
	payments app.PaymentRepository,
	contributions app.ContributionRepository,
	campaigns app.CampaignRepository,
	contributors app.ContributorRepository,
	organizations app.OrganizationRepository,
	email app.EmailSender,
	frontendURL string,
) *ConfirmService {
	return &ConfirmService{
		payments: payments, contributions: contributions, campaigns: campaigns,
		contributors: contributors, organizations: organizations,
		email: email, frontendURL: frontendURL,
	}
}

// HandleWebhookEvent processes an inbound provider webhook (Etapa 4 §5). Es
// idempotente: un evento repetido para un pago ya en el estado destino no
// tiene efecto doble (lo garantiza ConfirmByProviderRef). Eventos desconocidos
// son no-ops silenciosos — no amplificamos reintentos del proveedor con 4xx.
func (s *ConfirmService) HandleWebhookEvent(ctx context.Context, provider, providerRef, event string) error {
	var newStatus payment.Status
	switch event {
	case "payment.confirmed":
		newStatus = payment.StatusConfirmed
	case "payment.failed":
		newStatus = payment.StatusFailed
	default:
		return nil
	}
	confirmedPayment, transitioned, err := s.payments.ConfirmByProviderRef(ctx, provider, providerRef, newStatus)
	if err != nil {
		return err
	}
	if transitioned && newStatus == payment.StatusConfirmed {
		notifyOrganizer(ctx, s.contributions, s.campaigns, s.contributors, s.organizations, s.email, s.frontendURL, confirmedPayment)
	}
	return nil
}
