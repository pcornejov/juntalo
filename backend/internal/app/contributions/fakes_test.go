package contributions

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

// ── fakeCampaignRepo: solo lo que Start necesita (GetBySlug) ──────────────

type fakeCampaignRepo struct {
	bySlug map[string]campaign.Campaign
}

func newFakeCampaignRepo() *fakeCampaignRepo {
	return &fakeCampaignRepo{bySlug: map[string]campaign.Campaign{}}
}

func (f *fakeCampaignRepo) Create(context.Context, app.CreateCampaignInput) (campaign.Campaign, error) {
	return campaign.Campaign{}, nil
}
func (f *fakeCampaignRepo) GetByID(context.Context, uuid.UUID) (campaign.Campaign, bool, error) {
	return campaign.Campaign{}, false, nil
}
func (f *fakeCampaignRepo) GetByIDForOrg(context.Context, uuid.UUID, uuid.UUID) (campaign.Campaign, bool, error) {
	return campaign.Campaign{}, false, nil
}
func (f *fakeCampaignRepo) GetBySlug(_ context.Context, slug string) (campaign.Campaign, bool, error) {
	c, ok := f.bySlug[slug]
	return c, ok, nil
}
func (f *fakeCampaignRepo) SlugExists(context.Context, string) (bool, error) { return false, nil }
func (f *fakeCampaignRepo) ListByOrg(context.Context, uuid.UUID, int32, int32) ([]campaign.Campaign, error) {
	return nil, nil
}
func (f *fakeCampaignRepo) Update(context.Context, app.UpdateCampaignInput) (campaign.Campaign, error) {
	return campaign.Campaign{}, nil
}
func (f *fakeCampaignRepo) UpdateStatus(context.Context, uuid.UUID, campaign.Status) (campaign.Campaign, error) {
	return campaign.Campaign{}, nil
}
func (f *fakeCampaignRepo) SoftDelete(context.Context, uuid.UUID) error { return nil }
func (f *fakeCampaignRepo) GetTotals(context.Context, uuid.UUID) (campaign.Totals, error) {
	return campaign.Totals{}, nil
}
func (f *fakeCampaignRepo) PublishDueCampaigns(context.Context) ([]campaign.Campaign, error) {
	return nil, nil
}

func (f *fakeCampaignRepo) ListPublicByOrg(context.Context, uuid.UUID, int32, int32) ([]campaign.Campaign, error) {
	return nil, nil
}

func (f *fakeCampaignRepo) GetRaffleNumbersSold(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeCampaignRepo) ListPublic(context.Context, string, campaign.Category, int32, int32) ([]campaign.Campaign, error) {
	return nil, nil
}

func (f *fakeCampaignRepo) GetFeatured(context.Context) (campaign.Campaign, bool, error) {
	return campaign.Campaign{}, false, nil
}

// ── fakeOrgRepo ─────────────────────────────────────────────────────────

type fakeOrgRepo struct {
	byID map[uuid.UUID]float64 // commission rate
}

func (f *fakeOrgRepo) GetPersonalByUserID(context.Context, uuid.UUID) (identity.Organization, error) {
	return identity.Organization{}, nil
}
func (f *fakeOrgRepo) GetByID(_ context.Context, id uuid.UUID) (identity.Organization, error) {
	return identity.Organization{ID: id, CommissionRate: f.byID[id]}, nil
}
func (f *fakeOrgRepo) UpdateCommissionRate(context.Context, uuid.UUID, float64) (identity.Organization, bool, error) {
	return identity.Organization{}, false, nil
}

func (f *fakeOrgRepo) GetBySlug(context.Context, string) (identity.Organization, bool, error) {
	return identity.Organization{}, false, nil
}

func (f *fakeOrgRepo) GetOwnerEmail(context.Context, uuid.UUID) (string, string, error) {
	return "", "", nil
}
func (f *fakeOrgRepo) GetOwnerInfo(context.Context, uuid.UUID) (string, bool, error) {
	return "", false, nil
}

// ── fakeContributorRepo ─────────────────────────────────────────────────

type fakeContributorRepo struct{}

func (f *fakeContributorRepo) Create(context.Context, app.CreateContributorInput) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (f *fakeContributorRepo) GetByID(context.Context, uuid.UUID) (contribution.Contributor, bool, error) {
	return contribution.Contributor{}, false, nil
}

// ── fakeEmailSender: no-op para los tests que no verifican notificaciones ──

type fakeEmailSender struct{}

func (f *fakeEmailSender) Send(context.Context, string, string, string) error {
	return nil
}

// ── fakeContributionRepo ────────────────────────────────────────────────

type fakeContributionRepo struct {
	mu     sync.Mutex
	byID   map[uuid.UUID]contribution.Contribution
	nextID func() uuid.UUID
}

func newFakeContributionRepo() *fakeContributionRepo {
	return &fakeContributionRepo{byID: map[uuid.UUID]contribution.Contribution{}, nextID: uuid.New}
}

func (f *fakeContributionRepo) Create(_ context.Context, in app.CreateContributionInput) (contribution.Contribution, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := contribution.Contribution{
		ID:            f.nextID(),
		CampaignID:    in.CampaignID,
		ContributorID: in.ContributorID,
		Amount:        in.Amount,
		IsAnonymous:   in.IsAnonymous,
		Message:       in.Message,
		Status:        contribution.StatusPending,
	}
	f.byID[c.ID] = c
	return c, nil
}

func (f *fakeContributionRepo) CreateRaffleNumbered(_ context.Context, in app.CreateContributionInput, totalNumbers int) (contribution.Contribution, bool, error) {
	f.mu.Lock()
	sold := 0
	for _, c := range f.byID {
		if c.CampaignID == in.CampaignID && c.RaffleNumber != nil {
			sold++
		}
	}
	f.mu.Unlock()
	if sold >= totalNumbers {
		return contribution.Contribution{}, true, nil
	}
	num := sold + 1
	c, err := f.Create(context.Background(), in)
	if err != nil {
		return contribution.Contribution{}, false, err
	}
	f.mu.Lock()
	c.RaffleNumber = &num
	f.byID[c.ID] = c
	f.mu.Unlock()
	return c, false, nil
}

func (f *fakeContributionRepo) GetByID(_ context.Context, id uuid.UUID) (contribution.Contribution, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	return c, ok, nil
}

func (f *fakeContributionRepo) ListByCampaign(context.Context, uuid.UUID, int32, int32) ([]contribution.Contribution, error) {
	return nil, nil
}

func (f *fakeContributionRepo) setStatus(id uuid.UUID, status contribution.Status) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := f.byID[id]
	c.Status = status
	f.byID[id] = c
}

// ── fakePaymentRepo: el más importante — simula el UNIQUE(idempotency_key) ─

type fakePaymentRepo struct {
	mu            sync.Mutex
	byKey         map[string]*payment.Payment
	byProviderRef map[string]*payment.Payment
	contributions *fakeContributionRepo
}

func newFakePaymentRepo(contributions *fakeContributionRepo) *fakePaymentRepo {
	return &fakePaymentRepo{
		byKey:         map[string]*payment.Payment{},
		byProviderRef: map[string]*payment.Payment{},
		contributions: contributions,
	}
}

func (f *fakePaymentRepo) Create(_ context.Context, in app.CreatePaymentInput) (payment.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.byKey[in.IdempotencyKey]; exists {
		return payment.Payment{}, apperr.New("duplicate_contribution", "ya existe")
	}

	p := &payment.Payment{
		ID:                    uuid.New(),
		ContributionID:        in.ContributionID,
		IdempotencyKey:        in.IdempotencyKey,
		Provider:              in.Provider,
		ProviderRef:           in.ProviderRef,
		Status:                in.Status,
		AmountGross:           in.AmountGross,
		CommissionRateApplied: in.CommissionRateApplied,
		CommissionAmount:      in.CommissionAmount,
		AmountNet:             in.AmountNet,
	}
	f.byKey[in.IdempotencyKey] = p
	f.byProviderRef[in.Provider+"|"+in.ProviderRef] = p
	return *p, nil
}

func (f *fakePaymentRepo) GetByIdempotencyKey(_ context.Context, key string) (payment.Payment, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.byKey[key]
	if !ok {
		return payment.Payment{}, false, nil
	}
	return *p, true, nil
}

func (f *fakePaymentRepo) GetByContributionID(context.Context, uuid.UUID) (payment.Payment, bool, error) {
	return payment.Payment{}, false, nil
}

func (f *fakePaymentRepo) ConfirmByProviderRef(_ context.Context, provider, providerRef string, newStatus payment.Status) (payment.Payment, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	p, ok := f.byProviderRef[provider+"|"+providerRef]
	if !ok {
		return payment.Payment{}, false, apperr.New("payment_not_found", "no encontrado")
	}
	if p.Status == newStatus {
		return *p, false, nil // idempotente
	}
	if !payment.CanTransition(p.Status, newStatus) {
		return payment.Payment{}, false, apperr.New("invalid_payment_transition", "transición inválida")
	}
	p.Status = newStatus

	var contribStatus contribution.Status
	switch newStatus {
	case payment.StatusConfirmed:
		contribStatus = contribution.StatusConfirmed
	case payment.StatusFailed:
		contribStatus = contribution.StatusFailed
	default:
		contribStatus = contribution.StatusPending
	}
	f.contributions.setStatus(p.ContributionID, contribStatus)
	return *p, true, nil
}

func (f *fakePaymentRepo) Refund(context.Context, uuid.UUID, money.CLP, string, string) (payment.Payment, error) {
	return payment.Payment{}, apperr.New("payment_not_found", "no encontrado")
}

// ── fakeProvider: el mock del mock — controlado por el test ───────────────

type fakeProvider struct {
	mu      sync.Mutex
	mode    payment.Status // qué status devuelve CreateIntent
	intents map[string]string
}

func newFakeProvider(mode payment.Status) *fakeProvider {
	return &fakeProvider{mode: mode, intents: map[string]string{}}
}

func (f *fakeProvider) CreateIntent(_ context.Context, req app.IntentRequest) (app.PaymentIntent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if ref, ok := f.intents[req.IdempotencyKey]; ok {
		return app.PaymentIntent{ProviderRef: ref, Status: payment.StatusPending}, nil
	}
	ref := "fake_" + uuid.New().String()
	f.intents[req.IdempotencyKey] = ref
	return app.PaymentIntent{ProviderRef: ref, Status: f.mode}, nil
}

func (f *fakeProvider) GetIntent(context.Context, string) (app.PaymentIntent, error) {
	return app.PaymentIntent{}, nil
}

func (f *fakeProvider) Refund(context.Context, string, money.CLP) (app.RefundResult, error) {
	return app.RefundResult{}, nil
}
