// Package api wires the HTTP layer: Fiber app, middleware stack, and route table.
package api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/api/handlers"
	authuc "github.com/pcornejov/juntalo/backend/internal/app/auth"
	campaignsuc "github.com/pcornejov/juntalo/backend/internal/app/campaigns"
	filesuc "github.com/pcornejov/juntalo/backend/internal/app/files"
	infraauth "github.com/pcornejov/juntalo/backend/internal/infra/auth"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/repos"
	"github.com/pcornejov/juntalo/backend/internal/infra/storage/local"
)

type Config struct {
	JWTSecret   string
	IsProd      bool
	StorageDir  string
	StorageURL  string // URL pública base para archivos servidos localmente
	FrontendURL string // base de la SPA para el redirect de /c/:slug (Etapa 4 §4)
}

func NewServer(db *pgxpool.Pool, cfg Config) *fiber.App {
	fiberApp := fiber.New(fiber.Config{
		AppName: "juntalo-api",
	})

	fiberApp.Use(recover.New())
	fiberApp.Use(requestid.New())
	fiberApp.Use(logger.New())
	fiberApp.Use(cors.New())

	fiberApp.Get("/healthz", healthzHandler(db))
	fiberApp.Static("/files", cfg.StorageDir)

	hasher := infraauth.NewArgon2idHasher()
	signer := infraauth.NewJWTSigner(cfg.JWTSecret)
	storage := local.New(cfg.StorageDir, cfg.StorageURL)

	authRepo := repos.NewAuthRepo(db)
	userRepo := repos.NewUserRepo(db)
	orgRepo := repos.NewOrganizationRepo(db)
	refreshRepo := repos.NewRefreshTokenRepo(db)
	campaignRepo := repos.NewCampaignRepo(db)
	fileRepo := repos.NewFileRepo(db)

	registerSvc := authuc.NewRegisterService(authRepo, hasher)
	loginSvc := authuc.NewLoginService(userRepo, authRepo, orgRepo, hasher)
	refreshSvc := authuc.NewRefreshService(refreshRepo)

	createSvc := campaignsuc.NewCreateService(campaignRepo)
	getSvc := campaignsuc.NewGetService(campaignRepo)
	listSvc := campaignsuc.NewListService(campaignRepo)
	updateSvc := campaignsuc.NewUpdateService(campaignRepo)
	transitionSvc := campaignsuc.NewTransitionService(campaignRepo)
	deleteSvc := campaignsuc.NewDeleteService(campaignRepo)
	uploadSvc := filesuc.NewUploadService(storage, fileRepo)

	authHandler := handlers.NewAuthHandler(registerSvc, loginSvc, refreshSvc, userRepo, orgRepo, signer, cfg.IsProd)
	campaignHandler := handlers.NewCampaignHandler(createSvc, getSvc, listSvc, updateSvc, transitionSvc, deleteSvc, uploadSvc, orgRepo, fileRepo, storage)
	fileHandler := handlers.NewFileHandler(uploadSvc, orgRepo)
	publicHandler := handlers.NewPublicHandler(getSvc, fileRepo, storage, cfg.FrontendURL)

	v1 := fiberApp.Group("/api/v1")
	mountAuthRoutes(v1, authHandler, signer)
	mountCampaignRoutes(v1, campaignHandler, fileHandler, signer)
	mountPublicRoutes(fiberApp, v1, publicHandler)
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
