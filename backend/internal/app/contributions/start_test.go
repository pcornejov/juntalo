package contributions

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

type fixture struct {
	campaigns     *fakeCampaignRepo
	orgs          *fakeOrgRepo
	contributions *fakeContributionRepo
	payments      *fakePaymentRepo
	provider      *fakeProvider
	start         *StartService
	confirm       *ConfirmService
	campaignID    uuid.UUID
	orgID         uuid.UUID
}

func newFixture(t *testing.T, providerMode payment.Status) *fixture {
	t.Helper()
	orgID := uuid.New()
	campaignID := uuid.New()

	campaigns := newFakeCampaignRepo()
	campaigns.bySlug["ayuda-viaje"] = campaign.Campaign{
		ID: campaignID, OrganizationID: orgID, Status: campaign.StatusActive,
	}
	orgs := &fakeOrgRepo{byID: map[uuid.UUID]float64{orgID: 0.05}}
	contributionsRepo := newFakeContributionRepo()
	paymentsRepo := newFakePaymentRepo(contributionsRepo)
	provider := newFakeProvider(providerMode)

	return &fixture{
		campaigns: campaigns, orgs: orgs, contributions: contributionsRepo,
		payments: paymentsRepo, provider: provider,
		start:      NewStartService(campaigns, orgs, &fakeContributorRepo{}, contributionsRepo, paymentsRepo, provider),
		confirm:    NewConfirmService(paymentsRepo),
		campaignID: campaignID, orgID: orgID,
	}
}

func TestStartService_InstantConfirmation(t *testing.T) {
	f := newFixture(t, payment.StatusConfirmed)

	result, err := f.start.Start(context.Background(), StartInput{
		Slug: "ayuda-viaje", IdempotencyKey: "key-1", FullName: "Ana", Amount: 10_000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PaymentStatus != payment.StatusConfirmed {
		t.Errorf("got status %s, want confirmed", result.PaymentStatus)
	}

	c, found, _ := f.contributions.GetByID(context.Background(), result.ContributionID)
	if !found || c.Status != contribution.StatusConfirmed {
		t.Errorf("expected contribution confirmed, got found=%v status=%s", found, c.Status)
	}
}

func TestStartService_CommissionSnapshot(t *testing.T) {
	f := newFixture(t, payment.StatusConfirmed)

	result, err := f.start.Start(context.Background(), StartInput{
		Slug: "ayuda-viaje", IdempotencyKey: "key-1", FullName: "Ana", Amount: 100_000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pay, found, _ := f.payments.GetByIdempotencyKey(context.Background(), "key-1")
	if !found {
		t.Fatal("payment not found")
	}
	if pay.CommissionRateApplied != 0.05 {
		t.Errorf("got rate %v, want 0.05", pay.CommissionRateApplied)
	}
	if pay.CommissionAmount != 5_000 {
		t.Errorf("got commission %d, want 5000", pay.CommissionAmount)
	}
	if pay.AmountNet != 95_000 {
		t.Errorf("got net %d, want 95000", pay.AmountNet)
	}
	_ = result
}

// TestStartService_RepeatedIdempotencyKeyDoesNotDuplicate is el escenario
// central del Hito 3: reintentar el mismo submit no debe crear un segundo
// aporte ni cobrar dos veces.
func TestStartService_RepeatedIdempotencyKeyDoesNotDuplicate(t *testing.T) {
	f := newFixture(t, payment.StatusConfirmed)
	ctx := context.Background()

	first, err := f.start.Start(ctx, StartInput{Slug: "ayuda-viaje", IdempotencyKey: "same-key", FullName: "Ana", Amount: 10_000})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := f.start.Start(ctx, StartInput{Slug: "ayuda-viaje", IdempotencyKey: "same-key", FullName: "Ana", Amount: 10_000})
	if err != nil {
		t.Fatalf("second (retry): %v", err)
	}

	if first.ContributionID != second.ContributionID {
		t.Errorf("expected the same contribution on retry, got %v vs %v", first.ContributionID, second.ContributionID)
	}
	if len(f.payments.byKey) != 1 {
		t.Errorf("expected exactly 1 payment, got %d", len(f.payments.byKey))
	}
}

// TestStartService_ConcurrentSameKeyDoesNotDuplicate simula un doble-tap real:
// dos goroutines llaman Start al mismo tiempo con la misma Idempotency-Key.
func TestStartService_ConcurrentSameKeyDoesNotDuplicate(t *testing.T) {
	f := newFixture(t, payment.StatusConfirmed)
	ctx := context.Background()

	const attempts = 10
	var wg sync.WaitGroup
	results := make([]StartResult, attempts)
	errs := make([]error, attempts)

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r, err := f.start.Start(ctx, StartInput{Slug: "ayuda-viaje", IdempotencyKey: "race-key", FullName: "Ana", Amount: 10_000})
			results[i], errs[i] = r, err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("attempt %d failed: %v", i, err)
		}
	}
	firstID := results[0].ContributionID
	for i, r := range results {
		if r.ContributionID != firstID {
			t.Errorf("attempt %d got a different contribution: %v vs %v", i, r.ContributionID, firstID)
		}
	}
	if len(f.payments.byKey) != 1 {
		t.Errorf("expected exactly 1 payment despite %d concurrent attempts, got %d", attempts, len(f.payments.byKey))
	}
}

func TestStartService_CampaignNotActive(t *testing.T) {
	f := newFixture(t, payment.StatusConfirmed)
	f.campaigns.bySlug["ayuda-viaje"] = campaign.Campaign{ID: f.campaignID, OrganizationID: f.orgID, Status: campaign.StatusPaused}

	_, err := f.start.Start(context.Background(), StartInput{Slug: "ayuda-viaje", IdempotencyKey: "k", FullName: "Ana", Amount: 1000})
	if !apperr.Is(err, "campaign_not_active") {
		t.Fatalf("expected campaign_not_active, got %v", err)
	}
}

func TestStartService_UnknownCampaign(t *testing.T) {
	f := newFixture(t, payment.StatusConfirmed)

	_, err := f.start.Start(context.Background(), StartInput{Slug: "no-existe", IdempotencyKey: "k", FullName: "Ana", Amount: 1000})
	if !apperr.Is(err, "campaign_not_active") {
		t.Fatalf("expected campaign_not_active for unknown slug (no filtrar existencia), got %v", err)
	}
}

func TestStartService_InvalidAmount(t *testing.T) {
	f := newFixture(t, payment.StatusConfirmed)

	_, err := f.start.Start(context.Background(), StartInput{Slug: "ayuda-viaje", IdempotencyKey: "k", FullName: "Ana", Amount: 0})
	if !apperr.Is(err, "amount_out_of_range") {
		t.Fatalf("expected amount_out_of_range, got %v", err)
	}
}

// TestStartService_FailedPaymentDoesNotCountAsRaised: un aporte rechazado
// nunca debe confirmarse (Hito 3 demo: "un aporte fallido no suma").
func TestStartService_FailedPaymentDoesNotCountAsRaised(t *testing.T) {
	f := newFixture(t, payment.StatusFailed)

	result, err := f.start.Start(context.Background(), StartInput{Slug: "ayuda-viaje", IdempotencyKey: "k", FullName: "Ana", Amount: 10_000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PaymentStatus != payment.StatusFailed {
		t.Errorf("got status %s, want failed", result.PaymentStatus)
	}

	c, _, _ := f.contributions.GetByID(context.Background(), result.ContributionID)
	if c.Status != contribution.StatusFailed {
		t.Errorf("expected contribution failed, got %s", c.Status)
	}
}

// TestStartService_DeferredThenWebhookConfirms exercises the full async path:
// pending al crear, confirmado luego por el mismo mecanismo que usa el
// webhook real (ConfirmService.HandleWebhookEvent).
func TestStartService_DeferredThenWebhookConfirms(t *testing.T) {
	f := newFixture(t, payment.StatusPending)

	result, err := f.start.Start(context.Background(), StartInput{Slug: "ayuda-viaje", IdempotencyKey: "k", FullName: "Ana", Amount: 10_000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PaymentStatus != payment.StatusPending {
		t.Fatalf("expected pending right after start, got %s", result.PaymentStatus)
	}

	pay, _, _ := f.payments.GetByIdempotencyKey(context.Background(), "k")
	if err := f.confirm.HandleWebhookEvent(context.Background(), "mock", pay.ProviderRef, "payment.confirmed"); err != nil {
		t.Fatalf("webhook confirm: %v", err)
	}

	c, _, _ := f.contributions.GetByID(context.Background(), result.ContributionID)
	if c.Status != contribution.StatusConfirmed {
		t.Errorf("expected contribution confirmed after webhook, got %s", c.Status)
	}
}

// TestConfirmService_RepeatedWebhookEventIsIdempotent: un evento repetido no
// debe tener efecto doble (Etapa 4 §5).
func TestConfirmService_RepeatedWebhookEventIsIdempotent(t *testing.T) {
	f := newFixture(t, payment.StatusPending)

	result, err := f.start.Start(context.Background(), StartInput{Slug: "ayuda-viaje", IdempotencyKey: "k", FullName: "Ana", Amount: 10_000})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	pay, _, _ := f.payments.GetByIdempotencyKey(context.Background(), "k")

	if err := f.confirm.HandleWebhookEvent(context.Background(), "mock", pay.ProviderRef, "payment.confirmed"); err != nil {
		t.Fatalf("first webhook: %v", err)
	}
	if err := f.confirm.HandleWebhookEvent(context.Background(), "mock", pay.ProviderRef, "payment.confirmed"); err != nil {
		t.Fatalf("repeated webhook should be a no-op, not an error: %v", err)
	}

	c, _, _ := f.contributions.GetByID(context.Background(), result.ContributionID)
	if c.Status != contribution.StatusConfirmed {
		t.Errorf("expected confirmed, got %s", c.Status)
	}
}

func TestConfirmService_UnknownEventIsNoop(t *testing.T) {
	f := newFixture(t, payment.StatusPending)
	if _, err := f.start.Start(context.Background(), StartInput{Slug: "ayuda-viaje", IdempotencyKey: "k", FullName: "Ana", Amount: 10_000}); err != nil {
		t.Fatalf("start: %v", err)
	}
	pay, _, _ := f.payments.GetByIdempotencyKey(context.Background(), "k")

	if err := f.confirm.HandleWebhookEvent(context.Background(), "mock", pay.ProviderRef, "payment.unknown_event"); err != nil {
		t.Fatalf("unknown event should be a silent no-op, got %v", err)
	}
}
