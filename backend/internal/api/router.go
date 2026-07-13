package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/api/handlers"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
)

func mountAuthRoutes(router fiber.Router, h *handlers.AuthHandler, signer app.TokenSigner) {
	auth := router.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
	auth.Get("/me", middleware.RequireAuth(signer), h.Me)
}

func mountCampaignRoutes(router fiber.Router, h *handlers.CampaignHandler, fileH *handlers.FileHandler, signer app.TokenSigner) {
	campaigns := router.Group("/campaigns", middleware.RequireAuth(signer))
	campaigns.Get("/", h.List)
	campaigns.Post("/", h.Create)
	campaigns.Get("/:id", h.Get)
	campaigns.Patch("/:id", h.Update)
	campaigns.Delete("/:id", h.Delete)
	campaigns.Post("/:id/publish", h.Publish)
	campaigns.Post("/:id/pause", h.Pause)
	campaigns.Post("/:id/resume", h.Resume)
	campaigns.Post("/:id/finish", h.Finish)

	router.Post("/files", middleware.RequireAuth(signer), fileH.Upload)
}

func mountPublicRoutes(app fiber.Router, apiV1 fiber.Router, h *handlers.PublicHandler) {
	apiV1.Get("/public/campaigns/:slug", h.GetJSON)
	app.Get("/c/:slug", h.OGPage)
}

func mountMetaRoutes(router fiber.Router) {
	router.Get("/meta/campaign-types", handlers.CampaignTypes)
}
