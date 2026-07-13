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
