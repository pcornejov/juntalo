// Package api wires the HTTP layer: Fiber app, middleware stack, and route table.
package api

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/api/handlers"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
	adminuc "github.com/pcornejov/juntalo/backend/internal/app/admin"
	authuc "github.com/pcornejov/juntalo/backend/internal/app/auth"
	campaignsuc "github.com/pcornejov/juntalo/backend/internal/app/campaigns"
	contributionsuc "github.com/pcornejov/juntalo/backend/internal/app/contributions"
	dashboarduc "github.com/pcornejov/juntalo/backend/internal/app/dashboard"
	filesuc "github.com/pcornejov/juntalo/backend/internal/app/files"
	organizationsuc "github.com/pcornejov/juntalo/backend/internal/app/organizations"
	infraauth "github.com/pcornejov/juntalo/backend/internal/infra/auth"
	"github.com/pcornejov/juntalo/backend/internal/infra/captcha"
	"github.com/pcornejov/juntalo/backend/internal/infra/email"
	"github.com/pcornejov/juntalo/backend/internal/infra/payments/mock"
	"github.com/pcornejov/juntalo/backend/internal/infra/payments/webpay"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/repos"
	"github.com/pcornejov/juntalo/backend/internal/infra/storage/local"
	"github.com/pcornejov/juntalo/backend/internal/infra/storage/r2"
)

// publishSchedulerInterval: cada minuto es suficiente resolución para "en
// fecha programada" — no hace falta segundos de precisión para publicar
// campañas.
const publishSchedulerInterval = time.Minute

type Config struct {
	JWTSecret   string
	IsProd      bool
	StorageDir  string
	StorageURL  string // URL pública base para archivos servidos localmente
	FrontendURL string // base de la SPA para el redirect de /c/:slug (Etapa 4 §4)

	SelfURL           string // base propia para que el mock se autoinvoque vía webhook
	MockWebhookSecret string
	MockPaymentMode   string

	ExposeResetLinks bool // solo true en este deploy de prueba — ver Config.ExposeResetLinks en infra/config

	ResendAPIKey string
	EmailFrom    string

	// Storage persistente (Cloudflare R2). R2AccountID vacío = cae a
	// storage/local sobre StorageDir (conveniente en desarrollo; en
	// producción StorageDir es disco efímero, ver infra/config).
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	R2PublicURL       string

	// Pasarela de pago real (Webpay Plus). WebpayCommerceCode vacío = cae al
	// mock (ver infra/config).
	WebpayCommerceCode string
	WebpayAPIKey       string
	WebpayEnvironment  string

	// Backoffice del operador de la plataforma — string separado por comas,
	// vacío = nadie tiene acceso (ver infra/config).
	AdminEmails string

	// Captcha (Cloudflare Turnstile) en el registro. Vacío = sin captcha
	// (ver infra/config).
	TurnstileSecretKey string
}

func NewServer(db *pgxpool.Pool, cfg Config) *fiber.App {
	fiberApp := fiber.New(fiber.Config{
		AppName: "juntalo-api",
	})

	// Orden importa: recover.New() debe quedar afuera (red de seguridad
	// final que sí convierte el panic en una respuesta 500) y fibersentry
	// adentro, más cerca de los handlers, para que su propio recover()
	// vea el panic primero, lo reporte a Sentry, y recién ahí lo repropague
	// (Repanic: true) para que recover.New() termine el trabajo.
	fiberApp.Use(recover.New())
	fiberApp.Use(fibersentry.New(fibersentry.Config{Repanic: true}))
	fiberApp.Use(requestid.New())
	fiberApp.Use(logger.New())
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Idempotency-Key",
	}))
	// Auditoría de seguridad: el backend no seteaba ningún header de defensa
	// HTTP (el frontend sí los trae gratis de Render). CrossOriginResourcePolicy
	// se fuerza a "cross-origin" porque las imágenes de campaña se sirven desde
	// este dominio (juntalo-api) pero se embeben en <img> desde el dominio del
	// frontend (juntalo-web) — el default "same-origin" del middleware las
	// bloquearía. HSTS solo se activa en producción (Render sí sirve HTTPS).
	helmetCfg := helmet.Config{
		XFrameOptions:             "DENY",
		CrossOriginResourcePolicy: "cross-origin",
	}
	if cfg.IsProd {
		helmetCfg.HSTSMaxAge = 15552000 // 180 días
	}
	fiberApp.Use(helmet.New(helmetCfg))

	fiberApp.Get("/healthz", healthzHandler(db))
	fiberApp.Static("/files", cfg.StorageDir)

	hasher := infraauth.NewArgon2idHasher()
	signer := infraauth.NewJWTSigner(cfg.JWTSecret)
	var storage app.FileStorage
	if cfg.R2AccountID != "" {
		storage = r2.New(r2.Config{
			AccountID:       cfg.R2AccountID,
			AccessKeyID:     cfg.R2AccessKeyID,
			SecretAccessKey: cfg.R2SecretAccessKey,
			Bucket:          cfg.R2Bucket,
			PublicURL:       cfg.R2PublicURL,
		})
	} else {
		storage = local.New(cfg.StorageDir, cfg.StorageURL)
	}
	// Pasarela de pago: Webpay Plus (Transbank) si hay credenciales, si no
	// cae al mock — mismo patrón "vacío = off" que storage/R2 y Resend/Sentry.
	// webpayProvider queda como *webpay.Provider (no la interfaz) porque el
	// return handler necesita su método Commit, que no es parte de
	// app.PaymentProvider (ningún otro proveedor tiene ese paso).
	var paymentProvider app.PaymentProvider
	var webpayProvider *webpay.Provider
	var paymentProviderName string
	if cfg.WebpayCommerceCode != "" {
		webpayProvider = webpay.New(webpay.Config{
			CommerceCode: cfg.WebpayCommerceCode,
			APIKey:       cfg.WebpayAPIKey,
			Environment:  cfg.WebpayEnvironment,
			SelfURL:      cfg.SelfURL,
		})
		paymentProvider = webpayProvider
		paymentProviderName = "webpay"
	} else {
		paymentProvider = mock.NewProvider(
			mock.Mode(cfg.MockPaymentMode),
			cfg.SelfURL+"/api/v1/webhooks/payments/mock",
			cfg.MockWebhookSecret,
		)
		paymentProviderName = "mock"
	}
	var emailSender app.EmailSender
	if cfg.ResendAPIKey != "" {
		emailSender = email.NewResendSender(cfg.ResendAPIKey, cfg.EmailFrom)
	} else {
		emailSender = email.NewNoopSender()
	}
	var captchaVerifier app.CaptchaVerifier
	if cfg.TurnstileSecretKey != "" {
		captchaVerifier = captcha.NewTurnstileVerifier(cfg.TurnstileSecretKey)
	} else {
		captchaVerifier = captcha.NoopVerifier{}
	}

	authRepo := repos.NewAuthRepo(db)
	userRepo := repos.NewUserRepo(db)
	orgRepo := repos.NewOrganizationRepo(db)
	refreshRepo := repos.NewRefreshTokenRepo(db)
	campaignRepo := repos.NewCampaignRepo(db)
	fileRepo := repos.NewFileRepo(db)
	campaignImageRepo := repos.NewCampaignImageRepo(db)
	contributorRepo := repos.NewContributorRepo(db)
	contributionRepo := repos.NewContributionRepo(db)
	paymentRepo := repos.NewPaymentRepo(db)
	participantRepo := repos.NewParticipantRepo(db)
	auditRepo := repos.NewAuditRepo(db)
	passwordResetRepo := repos.NewPasswordResetRepo(db)
	emailVerificationRepo := repos.NewEmailVerificationRepo(db)
	adminRepo := repos.NewAdminRepo(db)
	payoutRepo := repos.NewPayoutRepo(db)

	adminEmails := splitAdminEmails(cfg.AdminEmails)

	registerSvc := authuc.NewRegisterService(authRepo, hasher, captchaVerifier)
	loginSvc := authuc.NewLoginService(userRepo, authRepo, orgRepo, hasher)
	refreshSvc := authuc.NewRefreshService(refreshRepo)
	forgotPasswordSvc := authuc.NewForgotPasswordService(userRepo, passwordResetRepo)
	resetPasswordSvc := authuc.NewResetPasswordService(passwordResetRepo, authRepo, hasher)
	emailVerifySvc := authuc.NewEmailVerificationService(userRepo, emailVerificationRepo)

	createSvc := campaignsuc.NewCreateService(campaignRepo)
	getSvc := campaignsuc.NewGetService(campaignRepo)
	listSvc := campaignsuc.NewListService(campaignRepo)
	updateSvc := campaignsuc.NewUpdateService(campaignRepo)
	transitionSvc := campaignsuc.NewTransitionService(campaignRepo)
	deleteSvc := campaignsuc.NewDeleteService(campaignRepo)
	cloneSvc := campaignsuc.NewCloneService(campaignRepo, createSvc)
	uploadSvc := filesuc.NewUploadService(storage, fileRepo)

	// Auto-publicar campaña en fecha programada: vive dentro de este mismo
	// proceso, no como infra de cron aparte (Etapa 4).
	go campaignsuc.RunPublishScheduler(context.Background(), campaignRepo, publishSchedulerInterval)

	startSvc := contributionsuc.NewStartService(campaignRepo, orgRepo, contributorRepo, contributionRepo, paymentRepo, paymentProvider, paymentProviderName, emailSender, cfg.FrontendURL)
	confirmSvc := contributionsuc.NewConfirmService(paymentRepo, contributionRepo, campaignRepo, contributorRepo, orgRepo, emailSender, cfg.FrontendURL)
	statusSvc := contributionsuc.NewStatusService(contributionRepo)
	participantsSvc := dashboarduc.NewParticipantsService(campaignRepo, participantRepo)
	exportSvc := dashboarduc.NewExportCSVService(campaignRepo, participantRepo)
	refundSvc := dashboarduc.NewRefundService(campaignRepo, contributionRepo, paymentRepo, paymentProvider)
	adminSvc := adminuc.NewService(adminRepo, campaignRepo, orgRepo, payoutRepo)
	orgProfileSvc := organizationsuc.NewService(orgRepo)

	authHandler := handlers.NewAuthHandler(registerSvc, loginSvc, refreshSvc, forgotPasswordSvc, resetPasswordSvc, emailVerifySvc, userRepo, orgRepo, signer, emailSender, cfg.FrontendURL, cfg.IsProd, cfg.ExposeResetLinks, adminEmails)
	campaignHandler := handlers.NewCampaignHandler(createSvc, getSvc, listSvc, updateSvc, transitionSvc, deleteSvc, cloneSvc, uploadSvc, orgRepo, fileRepo, campaignImageRepo, storage, auditRepo, cfg.SelfURL)
	dashboardHandler := handlers.NewDashboardHandler(participantsSvc, exportSvc, refundSvc, orgRepo)
	fileHandler := handlers.NewFileHandler(uploadSvc, orgRepo)
	publicHandler := handlers.NewPublicHandler(getSvc, listSvc, fileRepo, campaignImageRepo, storage, orgRepo, cfg.FrontendURL, cfg.SelfURL)
	contributionHandler := handlers.NewContributionHandler(startSvc, statusSvc)
	webhookHandler := handlers.NewWebhookHandler(confirmSvc, cfg.MockWebhookSecret)
	adminHandler := handlers.NewAdminHandler(adminSvc)
	organizationHandler := handlers.NewOrganizationHandler(orgProfileSvc, orgRepo)

	v1 := fiberApp.Group("/api/v1")
	mountAuthRoutes(v1, authHandler, signer)
	mountCampaignRoutes(v1, campaignHandler, dashboardHandler, fileHandler, signer)
	mountPublicRoutes(fiberApp, v1, publicHandler)
	mountContributionRoutes(v1, contributionHandler, middleware.ContributeLimiter())
	mountWebhookRoutes(v1, webhookHandler)
	mountMetaRoutes(v1)
	mountAdminRoutes(v1, adminHandler, signer, userRepo, adminEmails)
	mountOrganizationRoutes(v1, organizationHandler, signer)
	if webpayProvider != nil {
		webpayHandler := handlers.NewWebpayHandler(webpayProvider, confirmSvc, contributionRepo, campaignRepo, cfg.FrontendURL)
		mountWebpayRoutes(v1, webpayHandler)
	}

	return fiberApp
}

func splitAdminEmails(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func healthzHandler(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "down",
				"error":  err.Error(),
			})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
