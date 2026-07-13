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
	infraauth "github.com/pcornejov/juntalo/backend/internal/infra/auth"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/repos"
)

type Config struct {
	JWTSecret string
	IsProd    bool
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

	hasher := infraauth.NewArgon2idHasher()
	signer := infraauth.NewJWTSigner(cfg.JWTSecret)

	authRepo := repos.NewAuthRepo(db)
	userRepo := repos.NewUserRepo(db)
	orgRepo := repos.NewOrganizationRepo(db)
	refreshRepo := repos.NewRefreshTokenRepo(db)

	registerSvc := authuc.NewRegisterService(authRepo, hasher)
	loginSvc := authuc.NewLoginService(userRepo, authRepo, orgRepo, hasher)
	refreshSvc := authuc.NewRefreshService(refreshRepo)

	authHandler := handlers.NewAuthHandler(registerSvc, loginSvc, refreshSvc, userRepo, orgRepo, signer, cfg.IsProd)

	v1 := fiberApp.Group("/api/v1")
	mountAuthRoutes(v1, authHandler, signer)

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
