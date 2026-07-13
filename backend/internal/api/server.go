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
)

func NewServer(db *pgxpool.Pool) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "juntalo-api",
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/healthz", healthzHandler(db))

	api := app.Group("/api/v1")
	_ = api // rutas de Etapa 4 se montan en hitos siguientes

	return app
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
