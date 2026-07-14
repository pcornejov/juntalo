// Package api wires the HTTP layer: Fiber app, middleware stack, and route table.
package api

import (
	"context"
	"time"

	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/api/handlers"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
	authuc "github.com/pcornejov/juntalo/backend/internal/app/auth"
	campaignsuc "github.com/pcornejov/juntalo/backend/internal/app/campaigns"
	contributionsuc "github.com/pcornejov/juntalo/backend/internal/app/contributions"
	dashboarduc "github.com/pcornejov/juntalo/backend/internal/app/dashboard"
	filesuc "github.com/pcornejov/juntalo/backend/internal/app/files"
	infraauth "github.com/pcornejov/juntalo/backend/internal/infra/auth"
	"github.com/pcornejov/juntalo/backend/internal/infra/email"
	"github.com/pcornejov/juntalo/backend/internal/infra/payments/mock"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/repos"
	"github.com/pcornejov/juntalo/backend/internal/infra/storage/local"
)

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

	fiberApp.Get("/healthz", healthzHandler(db))
	fiberApp.Static("/files", cfg.StorageDir)

	hasher := infraauth.NewArgon2idHasher()
	signer := infraauth.NewJWTSigner(cfg.JWTSecret)
	storage := local.New(cfg.StorageDir, cfg.StorageURL)
	paymentProvider := mock.NewProvider(
		mock.Mode(cfg.MockPaymentMode),
		cfg.SelfURL+"/api/v1/webhooks/payments/mock",
		cfg.MockWebhookSecret,
	)
	var emailSender app.EmailSender
	if cfg.ResendAPIKey != "" {
		emailSender = email.NewResendSender(cfg.ResendAPIKey, cfg.EmailFrom)
	} else {
		emailSender = email.NewNoopSender()
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

	registerSvc := authuc.NewRegisterService(authRepo, hasher)
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
	uploadSvc := filesuc.NewUploadService(storage, fileRepo)

	startSvc := contributionsuc.NewStartService(campaignRepo, orgRepo, contributorRepo, contributionRepo, paymentRepo, paymentProvider, emailSender, cfg.FrontendURL)
	confirmSvc := contributionsuc.NewConfirmService(paymentRepo, contributionRepo, campaignRepo, contributorRepo, orgRepo, emailSender, cfg.FrontendURL)
	statusSvc := contributionsuc.NewStatusService(contributionRepo)
	participantsSvc := dashboarduc.NewParticipantsService(campaignRepo, participantRepo)
	exportSvc := dashboarduc.NewExportCSVService(campaignRepo, participantRepo)

	authHandler := handlers.NewAuthHandler(registerSvc, loginSvc, refreshSvc, forgotPasswordSvc, resetPasswordSvc, emailVerifySvc, userRepo, orgRepo, signer, emailSender, cfg.FrontendURL, cfg.IsProd, cfg.ExposeResetLinks)
	campaignHandler := handlers.NewCampaignHandler(createSvc, getSvc, listSvc, updateSvc, transitionSvc, deleteSvc, uploadSvc, orgRepo, fileRepo, campaignImageRepo, storage, auditRepo, cfg.SelfURL)
	dashboardHandler := handlers.NewDashboardHandler(participantsSvc, exportSvc, orgRepo)
	fileHandler := handlers.NewFileHandler(uploadSvc, orgRepo)
	publicHandler := handlers.NewPublicHandler(getSvc, fileRepo, campaignImageRepo, storage, cfg.FrontendURL, cfg.SelfURL)
	contributionHandler := handlers.NewContributionHandler(startSvc, statusSvc)
	webhookHandler := handlers.NewWebhookHandler(confirmSvc, cfg.MockWebhookSecret)

	v1 := fiberApp.Group("/api/v1")
	mountAuthRoutes(v1, authHandler, signer)
	mountCampaignRoutes(v1, campaignHandler, dashboardHandler, fileHandler, signer)
	mountPublicRoutes(fiberApp, v1, publicHandler)
	mountContributionRoutes(v1, contributionHandler, middleware.ContributeLimiter())
	mountWebhookRoutes(v1, webhookHandler)
	mountMetaRoutes(v1)

	return fiberApp
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
