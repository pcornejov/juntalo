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
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string) error
}

// PasswordResetRepository backs "olvidé mi contraseña": un token de un solo
// uso con expiración, igual en espíritu a RefreshTokenRepository pero sin
// necesidad de revocación explícita (se marca usado al consumirse).
type PasswordResetRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	// GetUserIDByValidHash solo devuelve resultado si el token no expiró y no
	// se usó antes — un token usado o vencido se trata como "no encontrado".
	GetUserIDByValidHash(ctx context.Context, tokenHash string) (uuid.UUID, bool, error)
	MarkUsed(ctx context.Context, tokenHash string) error
}

// EmailVerificationRepository tiene exactamente la misma forma que
// PasswordResetRepository — mismo patrón de token de un solo uso, distinto
// propósito. Se mantiene como interfaz/tabla separada (igual que
// refresh_tokens vs password_reset_tokens) para no acoplar dos flujos que
// evolucionan distinto.
type EmailVerificationRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetUserIDByValidHash(ctx context.Context, tokenHash string) (uuid.UUID, bool, error)
	MarkUsed(ctx context.Context, tokenHash string) error
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (identity.User, bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (identity.User, bool, error)
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
}

type OrganizationRepository interface {
	GetPersonalByUserID(ctx context.Context, userID uuid.UUID) (identity.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (identity.Organization, error)
	// GetBySlug resuelve la página pública persistente del organizador
	// (/org/:slug) — inspirada en el link único de por vida de Ceneka.
	GetBySlug(ctx context.Context, slug string) (identity.Organization, bool, error)
	// UpdateCommissionRate: herramienta del backoffice para ajustar la
	// comisión de una organización sin tocar código (rate como fracción,
	// ej. 0.05 = 5%). found=false si el id no existe.
	UpdateCommissionRate(ctx context.Context, id uuid.UUID, rate float64) (identity.Organization, bool, error)
	// GetOwnerEmail se usa para notificar al organizador de nuevos aportes
	// (Hito "notificaciones") — no requiere UI de equipos, solo el dueño.
	GetOwnerEmail(ctx context.Context, id uuid.UUID) (email, fullName string, err error)
	// GetOwnerInfo agrega IsVerified (email del dueño verificado) sobre
	// GetOwnerEmail — proxy honesto para el badge de verificación de la
	// página pública (inspirado en Vaki), sin cola de revisión manual.
	GetOwnerInfo(ctx context.Context, id uuid.UUID) (fullName string, isVerified bool, err error)
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

// EmailSender isolates the transactional email provider (Resend hoy) del
// resto del código — mismo espíritu que PaymentProvider: una interfaz
// estable, implementaciones intercambiables (Resend real, o un no-op
// cuando no hay API key configurada, para no romper el flujo si el envío
// de email todavía no está prendido en un ambiente).
type EmailSender interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

// TokenSigner issues and parses short-lived JWT access tokens.
type TokenSigner interface {
	Sign(userID uuid.UUID, ttl time.Duration) (string, error)
	Parse(token string) (uuid.UUID, error)
}

// CaptchaVerifier valida un token de Cloudflare Turnstile antes de crear una
// cuenta — mismo espíritu que EmailSender: un no-op siempre-true cuando no
// hay TURNSTILE_SECRET_KEY configurada, para no romper el flujo en
// desarrollo/test.
type CaptchaVerifier interface {
	Verify(ctx context.Context, token, remoteIP string) (bool, error)
}

// CreateCampaignInput carries everything CampaignRepository.Create needs.
type CreateCampaignInput struct {
	OrganizationID uuid.UUID
	TypeKey        campaign.TypeKey
	Category       campaign.Category
	Title          string
	Slug           string
	Description    string
	GoalAmount     *money.CLP
	StartsAt       *time.Time
	EndsAt         *time.Time
	// PublishAt: si viene seteada, la campaña queda en draft hasta que el
	// scheduler en background la publique automáticamente (Etapa 4).
	PublishAt          *time.Time
	VideoURL           *string
	RaffleUnitPrice    *money.CLP
	RaffleTotalNumbers *int
}

// UpdateCampaignInput carries the editable fields of a campaign (Etapa 4 §3).
type UpdateCampaignInput struct {
	ID                  uuid.UUID
	Title               string
	Description         string
	GoalAmount          *money.CLP
	StartsAt            *time.Time
	EndsAt              *time.Time
	CoverFileID         *uuid.UUID
	PublishAt           *time.Time
	Category            campaign.Category
	VideoURL            *string
	RaffleUnitPrice     *money.CLP
	RaffleTotalNumbers  *int
	RaffleWinningNumber *int
}

type CampaignRepository interface {
	Create(ctx context.Context, in CreateCampaignInput) (campaign.Campaign, error)
	GetByID(ctx context.Context, id uuid.UUID) (campaign.Campaign, bool, error)
	GetByIDForOrg(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, bool, error)
	GetBySlug(ctx context.Context, slug string) (campaign.Campaign, bool, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]campaign.Campaign, error)
	// ListPublic lista campañas activas de cualquier organización, para la
	// sección pública de "Campañas activas" (Etapa 4) — search/category
	// vacíos desactivan cada filtro.
	ListPublic(ctx context.Context, search string, category campaign.Category, limit, offset int32) ([]campaign.Campaign, error)
	// ListPublicByOrg lista las campañas visibles públicamente (active/paused/
	// finished — misma política que GetPublicBySlug) de un organizador, para
	// su página de perfil persistente (/org/:slug).
	ListPublicByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]campaign.Campaign, error)
	// GetRaffleNumbersSold cuenta los números reservados/vendidos (pending +
	// confirmed) de una campaña de tipo "raffle" — total_numbers menos esto
	// es lo que queda disponible para la venta.
	GetRaffleNumbersSold(ctx context.Context, campaignID uuid.UUID) (int64, error)
	// GetFeatured devuelve la campaña activa con el aporte confirmado más
	// reciente — la "más caliente" (inspirado en Vaki), found=false si
	// ninguna campaña activa tiene aportes confirmados todavía.
	GetFeatured(ctx context.Context) (campaign.Campaign, bool, error)
	Update(ctx context.Context, in UpdateCampaignInput) (campaign.Campaign, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status campaign.Status) (campaign.Campaign, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetTotals(ctx context.Context, id uuid.UUID) (campaign.Totals, error)
	// PublishDueCampaigns transiciona atómicamente a 'active' todo draft cuya
	// publish_at ya venció — usado por el scheduler en background (Etapa 4).
	PublishDueCampaigns(ctx context.Context) ([]campaign.Campaign, error)
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

// CampaignImageRecord is a gallery image already resolved to its storage
// key, in display order, for a single campaign.
type CampaignImageRecord struct {
	ID         uuid.UUID
	StorageKey string
}

// CampaignImageRepository backs the campaign photo gallery/carousel: varias
// imágenes por campaña en vez de una sola cover_file_id (pedido del
// producto tras el MVP inicial).
type CampaignImageRepository interface {
	// Add appends fileID to campaignID's gallery, at the next position.
	Add(ctx context.Context, campaignID, fileID uuid.UUID) (uuid.UUID, error)
	ListByCampaign(ctx context.Context, campaignID uuid.UUID) ([]CampaignImageRecord, error)
	// Delete removes imageID from campaignID's gallery; returns false if it
	// didn't belong to that campaign (so handlers can 404 instead of
	// silently no-op-ing on someone else's image id).
	Delete(ctx context.Context, campaignID, imageID uuid.UUID) (bool, error)
	// Reorder rewrites position 0..n-1 following orderedImageIDs. La primera
	// imagen queda como portada — "elegir portada" es simplemente moverla al
	// principio. Falla si algún id no pertenece a campaignID (evita que un
	// organizador reordene fotos de una campaña ajena colándolas en el body).
	Reorder(ctx context.Context, campaignID uuid.UUID, orderedImageIDs []uuid.UUID) error
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
	GetByID(ctx context.Context, id uuid.UUID) (contribution.Contributor, bool, error)
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
	// CreateRaffleNumbered asigna atómicamente el siguiente número de rifa
	// disponible (bloqueando la fila de la campaña durante la transacción,
	// mismo patrón que las transacciones financieras de PaymentRepo) y crea
	// la contribución con ese número. soldOut=true si ya no quedan números
	// disponibles dentro de totalNumbers.
	CreateRaffleNumbered(ctx context.Context, in CreateContributionInput, totalNumbers int) (result contribution.Contribution, soldOut bool, err error)
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
	// transitioned indica si esta llamada realmente cambió el estado (false
	// si el pago ya estaba en newStatus) — el caller lo usa para no mandar
	// una notificación duplicada cuando el proveedor reintenta el webhook.
	ConfirmByProviderRef(ctx context.Context, provider, providerRef string, newStatus payment.Status) (result payment.Payment, transitioned bool, err error)
	// Refund registra un reembolso (total o parcial) sobre paymentID en una
	// transacción: inserta el registro en payment_refunds y transiciona el
	// pago (y su contribution vinculada) a partially_refunded o refunded
	// según si amount cubre o no el saldo pendiente (Etapa 3 §5).
	Refund(ctx context.Context, paymentID uuid.UUID, amount money.CLP, providerRef, reason string) (payment.Payment, error)
}

// ── Panel del organizador (Hito 4) ──────────────────────────────────────────

// ParticipantRow is the dashboard read model: un aporte con los datos del
// contribuyente y lo reembolsado, para participantes y export CSV
// (Etapa 1 riesgo 10: los reembolsos restan en el reporting desde el día 1).
type ParticipantRow struct {
	ContributionID uuid.UUID
	FullName       string
	Email          string
	Phone          string
	Amount         money.CLP
	RefundedAmount money.CLP
	IsAnonymous    bool
	Status         contribution.Status
	CreatedAt      time.Time
	Message        string
	// RaffleNumber: número asignado cuando la campaña es de tipo "raffle" —
	// nil para el resto de los tipos. Es lo que el organizador necesita para
	// identificar al ganador tras el sorteo externo.
	RaffleNumber *int
}

type ParticipantRepository interface {
	ListByCampaign(ctx context.Context, campaignID uuid.UUID, limit, offset int32) ([]ParticipantRow, error)
	// ListByCampaignFiltered scopes ListByCampaign con búsqueda (nombre/email/
	// teléfono) y filtro por estado — search/status vacíos desactivan cada
	// filtro (QA: el buscador del dashboard solo filtraba client-side sobre
	// la página cargada, dando falsos negativos en campañas con >1 página).
	ListByCampaignFiltered(ctx context.Context, campaignID uuid.UUID, search, status string, limit, offset int32) ([]ParticipantRow, error)
}

// RecordAuditInput carries what AuditRepository.Record persists (Etapa 4 §1,
// Etapa 5: auditoría enfocada en cambios de estado de campaña).
type RecordAuditInput struct {
	ActorUserID    uuid.UUID
	OrganizationID uuid.UUID
	Action         string
	EntityType     string
	EntityID       uuid.UUID
	Data           map[string]any
}

type AuditRepository interface {
	Record(ctx context.Context, in RecordAuditInput) error
}

// ── Backoffice (operador de la plataforma) ──────────────────────────────────

// AdminUserRow is one row of the platform-wide user list — a cuenta y su
// organización personal, con cuántas campañas tiene.
type AdminUserRow struct {
	ID                         uuid.UUID
	Email                      string
	FullName                   string
	EmailVerified              bool
	CreatedAt                  time.Time
	OrganizationID             uuid.UUID
	OrganizationName           string
	OrganizationCommissionRate float64
	CampaignCount              int64
}

// AdminCampaignRow is one row of the platform-wide campaign list — a
// diferencia de CampaignWithTotals (Etapa "dashboard del organizador"),
// incluye quién la organiza porque no está scopeada a una sola organización.
type AdminCampaignRow struct {
	ID               uuid.UUID
	Title            string
	Slug             string
	TypeKey          campaign.TypeKey
	Status           campaign.Status
	Category         campaign.Category
	CreatedAt        time.Time
	OrganizationName string
	OrganizerEmail   string
	RaisedGross      money.CLP
	ContributorCount int64
}

// AdminPaymentRow is one row of the platform-wide payment list — para
// auditar transacciones sin entrar campaña por campaña.
type AdminPaymentRow struct {
	ID               uuid.UUID
	Status           payment.Status
	Provider         string
	AmountGross      money.CLP
	AmountNet        money.CLP
	CommissionAmount money.CLP
	CreatedAt        time.Time
	ConfirmedAt      *time.Time
	CampaignTitle    string
	CampaignSlug     string
	ContributorName  string
}

// AdminMetrics is the platform-wide summary shown at the top of the
// backoffice — RaisedGross/RaisedNetApprox se agregan sobre campaign_totals
// (que ya descuenta reembolsos), no sobre payments directo.
type AdminMetrics struct {
	TotalUsers         int64
	TotalCampaigns     int64
	ActiveCampaigns    int64
	DraftCampaigns     int64
	FinishedCampaigns  int64
	TotalContributions int64
	RaisedGross        money.CLP
	RaisedNetApprox    money.CLP
	TotalCommission    money.CLP
}

// AdminRepository queries cross-tenant, a diferencia de todo el resto del
// código — solo lo puede llamar un admin de plataforma (ver
// middleware.RequireAdminUser), nunca un handler autenticado normal.
type AdminRepository interface {
	ListUsers(ctx context.Context, limit, offset int32) ([]AdminUserRow, error)
	ListCampaigns(ctx context.Context, limit, offset int32) ([]AdminCampaignRow, error)
	ListPayments(ctx context.Context, limit, offset int32) ([]AdminPaymentRow, error)
	GetMetrics(ctx context.Context) (AdminMetrics, error)
}
