// Package contributions implements the public contribution flow: the most
// financially sensitive part of Juntalo (Etapa 5: "concentran el riesgo
// financiero del producto; son los más testeados del repo").
package contributions

import (
	"context"
	"strconv"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

var errCampaignNotActive = apperr.New("campaign_not_active", "Esta campaña no está recibiendo aportes en este momento")

type StartService struct {
	campaigns     app.CampaignRepository
	organizations app.OrganizationRepository
	contributors  app.ContributorRepository
	contributions app.ContributionRepository
	payments      app.PaymentRepository
	provider      app.PaymentProvider
	// providerName se persiste en payments.provider y es la clave con la que
	// después se busca el pago por provider_ref (ConfirmByProviderRef, tanto
	// acá mismo como en el webhook/return handler de cada proveedor) — tiene
	// que coincidir exacto con el nombre que usa ese proveedor para
	// confirmar, o la confirmación nunca encuentra el pago (bug real: quedó
	// hardcodeado en "mock" hasta que se integró Webpay, momento en el que
	// toda confirmación real fallaba en silencio por buscar provider="mock"
	// contra filas guardadas con el provider correcto).
	providerName string
	email        app.EmailSender
	frontendURL  string
}

func NewStartService(
	campaigns app.CampaignRepository,
	organizations app.OrganizationRepository,
	contributors app.ContributorRepository,
	contributions app.ContributionRepository,
	payments app.PaymentRepository,
	provider app.PaymentProvider,
	providerName string,
	email app.EmailSender,
	frontendURL string,
) *StartService {
	return &StartService{
		campaigns: campaigns, organizations: organizations, contributors: contributors,
		contributions: contributions, payments: payments, provider: provider,
		providerName: providerName, email: email, frontendURL: frontendURL,
	}
}

type StartInput struct {
	Slug           string
	IdempotencyKey string
	FullName       string
	Email          string
	Phone          string
	Amount         money.CLP
	IsAnonymous    bool
	Message        string
}

type StartResult struct {
	ContributionID uuid.UUID
	PaymentStatus  payment.Status
	RedirectURL    string
}

// Start implements the flow of Etapa 4 §4: valida campaña activa → crea
// contributor + contribution(pending) + payment(pending) con snapshot de
// comisión → CreateIntent. Es idempotente en IdempotencyKey de punta a punta.
func (s *StartService) Start(ctx context.Context, in StartInput) (StartResult, error) {
	if existing, found, err := s.payments.GetByIdempotencyKey(ctx, in.IdempotencyKey); err != nil {
		return StartResult{}, err
	} else if found {
		return StartResult{ContributionID: existing.ContributionID, PaymentStatus: existing.Status}, nil
	}

	if err := contribution.ValidateAmount(in.Amount); err != nil {
		return StartResult{}, err
	}
	if err := contribution.ValidateContributorName(in.FullName); err != nil {
		return StartResult{}, err
	}

	camp, found, err := s.campaigns.GetBySlug(ctx, in.Slug)
	if err != nil {
		return StartResult{}, err
	}
	if !found || camp.Status != campaign.StatusActive {
		return StartResult{}, errCampaignNotActive
	}

	org, err := s.organizations.GetByID(ctx, camp.OrganizationID)
	if err != nil {
		return StartResult{}, err
	}

	contributorID, err := s.contributors.Create(ctx, app.CreateContributorInput{
		FullName: in.FullName,
		Email:    in.Email,
		Phone:    in.Phone,
	})
	if err != nil {
		return StartResult{}, err
	}

	newContribution, err := s.contributions.Create(ctx, app.CreateContributionInput{
		CampaignID:    camp.ID,
		ContributorID: contributorID,
		Amount:        in.Amount,
		IsAnonymous:   in.IsAnonymous,
		Message:       in.Message,
	})
	if err != nil {
		return StartResult{}, err
	}

	commissionAmount, netAmount := payment.ComputeCommission(in.Amount, org.CommissionRate)
	payeeSnapshot := map[string]string{
		"organization_id": camp.OrganizationID.String(),
		"commission_rate": strconv.FormatFloat(org.CommissionRate, 'f', 4, 64),
	}

	intent, err := s.provider.CreateIntent(ctx, app.IntentRequest{
		IdempotencyKey: in.IdempotencyKey,
		Amount:         in.Amount,
		Commission:     commissionAmount,
		PayeeAccount:   payeeSnapshot,
	})
	if err != nil {
		return StartResult{}, err
	}

	// El pago SIEMPRE nace pending; ConfirmByProviderRef es el único mecanismo
	// que lo transiciona (sea aquí mismo, si el provider confirma/rechaza al
	// instante, o después vía webhook) — así nunca hay dos caminos que muevan
	// el mismo estado.
	_, err = s.payments.Create(ctx, app.CreatePaymentInput{
		ContributionID:        newContribution.ID,
		IdempotencyKey:        in.IdempotencyKey,
		Provider:              s.providerName,
		ProviderRef:           intent.ProviderRef,
		Status:                payment.StatusPending,
		AmountGross:           in.Amount,
		CommissionRateApplied: org.CommissionRate,
		CommissionAmount:      commissionAmount,
		AmountNet:             netAmount,
		PayeeSnapshot:         payeeSnapshot,
	})
	if err != nil {
		if apperr.Is(err, "duplicate_contribution") {
			// Doble submit concurrente con la misma key: la primera tx ganó la
			// carrera; devolvemos su resultado en vez de propagar el conflicto.
			existing, found, getErr := s.payments.GetByIdempotencyKey(ctx, in.IdempotencyKey)
			if getErr != nil {
				return StartResult{}, getErr
			}
			if found {
				return StartResult{ContributionID: existing.ContributionID, PaymentStatus: existing.Status}, nil
			}
		}
		return StartResult{}, err
	}

	finalStatus := payment.StatusPending
	if intent.Status == payment.StatusConfirmed || intent.Status == payment.StatusFailed {
		confirmedPayment, transitioned, err := s.payments.ConfirmByProviderRef(ctx, s.providerName, intent.ProviderRef, intent.Status)
		if err != nil {
			return StartResult{}, err
		}
		finalStatus = intent.Status
		if transitioned && intent.Status == payment.StatusConfirmed {
			notifyOrganizer(ctx, s.contributions, s.campaigns, s.contributors, s.organizations, s.email, s.frontendURL, confirmedPayment)
		}
	}

	return StartResult{
		ContributionID: newContribution.ID,
		PaymentStatus:  finalStatus,
		RedirectURL:    intent.RedirectURL,
	}, nil
}
