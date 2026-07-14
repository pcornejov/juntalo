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
// Devuelve el payment actualizado (zero value si el evento fue un no-op) para
// que un caller que necesite más contexto — ej. el return handler de Webpay,
// que arma el redirect final al frontend a partir del contribution_id — no
// tenga que volver a consultarlo por su cuenta.
func (s *ConfirmService) HandleWebhookEvent(ctx context.Context, provider, providerRef, event string) (payment.Payment, error) {
	var newStatus payment.Status
	switch event {
	case "payment.confirmed":
		newStatus = payment.StatusConfirmed
	case "payment.failed":
		newStatus = payment.StatusFailed
	default:
		return payment.Payment{}, nil
	}
	confirmedPayment, transitioned, err := s.payments.ConfirmByProviderRef(ctx, provider, providerRef, newStatus)
	if err != nil {
		return payment.Payment{}, err
	}
	if transitioned && newStatus == payment.StatusConfirmed {
		notifyOrganizer(ctx, s.contributions, s.campaigns, s.contributors, s.organizations, s.email, s.frontendURL, confirmedPayment)
	}
	return confirmedPayment, nil
}
