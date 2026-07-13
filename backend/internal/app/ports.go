// Package app contains use cases (casos de uso) that orchestrate domain logic and
// infrastructure ports. All repository/adapter interfaces live in this single file
// (Etapa 5 §2: "se ve el contrato de persistencia completo de un vistazo").
package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

// RegisterInput carries everything AuthRepository.Register needs to run the
// user+identity+organization+membership transaction (Etapa 3 §2).
type RegisterInput struct {
	Email        string
	FullName     string
	PasswordHash string
}

type AuthRepository interface {
	// Register creates user + password identity + personal organization + owner
	// membership in a single transaction. Returns apperr "email_already_registered"
	// if the email is taken.
	Register(ctx context.Context, in RegisterInput) (identity.User, identity.Organization, error)
	GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error)
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (identity.User, bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (identity.User, bool, error)
}

type OrganizationRepository interface {
	GetPersonalByUserID(ctx context.Context, userID uuid.UUID) (identity.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (identity.Organization, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t identity.RefreshToken) error
	GetValidByHash(ctx context.Context, tokenHash string) (identity.RefreshToken, bool, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}

// PasswordHasher isolates the hashing algorithm (argon2id) from the use cases that need it.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// TokenSigner issues and parses short-lived JWT access tokens.
type TokenSigner interface {
	Sign(userID uuid.UUID, ttl time.Duration) (string, error)
	Parse(token string) (uuid.UUID, error)
}

// CreateCampaignInput carries everything CampaignRepository.Create needs.
type CreateCampaignInput struct {
	OrganizationID uuid.UUID
	TypeKey        campaign.TypeKey
	Title          string
	Slug           string
	Description    string
	GoalAmount     *money.CLP
	StartsAt       *time.Time
	EndsAt         *time.Time
}

// UpdateCampaignInput carries the editable fields of a campaign (Etapa 4 §3).
type UpdateCampaignInput struct {
	ID          uuid.UUID
	Title       string
	Description string
	GoalAmount  *money.CLP
	StartsAt    *time.Time
	EndsAt      *time.Time
	CoverFileID *uuid.UUID
}

type CampaignRepository interface {
	Create(ctx context.Context, in CreateCampaignInput) (campaign.Campaign, error)
	GetByID(ctx context.Context, id uuid.UUID) (campaign.Campaign, bool, error)
	GetByIDForOrg(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, bool, error)
	GetBySlug(ctx context.Context, slug string) (campaign.Campaign, bool, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]campaign.Campaign, error)
	Update(ctx context.Context, in UpdateCampaignInput) (campaign.Campaign, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status campaign.Status) (campaign.Campaign, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetTotals(ctx context.Context, id uuid.UUID) (campaign.Totals, error)
}

// CreateFileInput carries what FileRepository.Create needs to persist file metadata
// after the bytes are already written to storage (Etapa 4 §6).
type CreateFileInput struct {
	OrganizationID uuid.UUID
	Kind           string
	StorageKey     string
	MimeType       string
	SizeBytes      int64
}

type FileRecord struct {
	ID         uuid.UUID
	StorageKey string
	MimeType   string
}

type FileRepository interface {
	Create(ctx context.Context, in CreateFileInput) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (FileRecord, bool, error)
}

// FileStorage isolates where file bytes actually live: local disk today, R2
// tomorrow (Etapa 2 §2.1) — same interface, different adapter.
type FileStorage interface {
	Save(ctx context.Context, key string, data []byte) error
	PublicURL(key string) string
}

// ── Pagos y contribuciones (Hito 3 — el corazón del producto) ──────────────

// IntentRequest carries what PaymentProvider.CreateIntent needs. Commission and
// PayeeAccount are declared at intent-creation time because split payment
// requires that (Etapa 2 §2.3): la comisión se declara al crear el intento.
type IntentRequest struct {
	IdempotencyKey string
	Amount         money.CLP
	Commission     money.CLP
	PayeeAccount   map[string]string
}

type PaymentIntent struct {
	ProviderRef string
	Status      payment.Status // pending, confirmed o failed al crearse
	RedirectURL string
}

type RefundResult struct {
	ProviderRef string
	Amount      money.CLP
}

// PaymentProvider is the abstraction over payment gateways (Etapa 2 §2.3).
// MockPaymentProvider implements it today; Webpay/Mercado Pago/Stripe/Khipu
// implement it tomorrow without touching any use case.
type PaymentProvider interface {
	CreateIntent(ctx context.Context, req IntentRequest) (PaymentIntent, error)
	GetIntent(ctx context.Context, providerRef string) (PaymentIntent, error)
	Refund(ctx context.Context, providerRef string, amount money.CLP) (RefundResult, error)
}

type CreateContributorInput struct {
	FullName string
	Email    string
	Phone    string
}

type ContributorRepository interface {
	Create(ctx context.Context, in CreateContributorInput) (uuid.UUID, error)
}

type CreateContributionInput struct {
	CampaignID    uuid.UUID
	ContributorID uuid.UUID
	Amount        money.CLP
	IsAnonymous   bool
	Message       string
}

type ContributionRepository interface {
	Create(ctx context.Context, in CreateContributionInput) (contribution.Contribution, error)
	GetByID(ctx context.Context, id uuid.UUID) (contribution.Contribution, bool, error)
	ListByCampaign(ctx context.Context, campaignID uuid.UUID, limit, offset int32) ([]contribution.Contribution, error)
}

// CreatePaymentInput carries the full financial snapshot taken at contribution
// time (Etapa 3 §5, riesgo 3): commission rate/amount never recomputed later.
type CreatePaymentInput struct {
	ContributionID        uuid.UUID
	IdempotencyKey        string
	Provider              string
	ProviderRef           string
	Status                payment.Status
	AmountGross           money.CLP
	CommissionRateApplied float64
	CommissionAmount      money.CLP
	AmountNet             money.CLP
	PayeeSnapshot         map[string]string
}

type PaymentRepository interface {
	Create(ctx context.Context, in CreatePaymentInput) (payment.Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (payment.Payment, bool, error)
	GetByContributionID(ctx context.Context, contributionID uuid.UUID) (payment.Payment, bool, error)
	// ConfirmByProviderRef atomically transitions the payment (found by provider
	// + providerRef) and its linked contribution in a single transaction
	// (Etapa 4 §5: idempotente — un evento repetido no debe tener efecto doble).
	ConfirmByProviderRef(ctx context.Context, provider, providerRef string, newStatus payment.Status) error
}
