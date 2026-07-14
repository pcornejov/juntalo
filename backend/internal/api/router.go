package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/api/handlers"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
)

// mountAuthRoutes: register/login/forgot-password/reset-password llevan
// AuthLimiter (auditoría de seguridad: sin esto, fuerza bruta de contraseñas
// y creación masiva de cuentas quedaban sin fricción alguna).
func mountAuthRoutes(router fiber.Router, h *handlers.AuthHandler, signer app.TokenSigner) {
	authLimiter := middleware.AuthLimiter()
	auth := router.Group("/auth")
	auth.Post("/register", authLimiter, h.Register)
	auth.Post("/login", authLimiter, h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
	auth.Post("/forgot-password", authLimiter, h.ForgotPassword)
	auth.Post("/reset-password", authLimiter, h.ResetPassword)
	auth.Post("/verify-email", h.VerifyEmail)
	auth.Post("/resend-verification", middleware.RequireAuth(signer), h.ResendVerification)
	auth.Get("/me", middleware.RequireAuth(signer), h.Me)
}

func mountCampaignRoutes(router fiber.Router, h *handlers.CampaignHandler, dashH *handlers.DashboardHandler, fileH *handlers.FileHandler, signer app.TokenSigner) {
	campaigns := router.Group("/campaigns", middleware.RequireAuth(signer))
	campaigns.Get("/", h.List)
	campaigns.Post("/", h.Create)
	campaigns.Get("/:id", h.Get)
	campaigns.Patch("/:id", h.Update)
	campaigns.Delete("/:id", h.Delete)
	campaigns.Post("/:id/clone", h.Clone)
	campaigns.Post("/:id/cancel-schedule", h.CancelSchedule)
	campaigns.Post("/:id/publish", h.Publish)
	campaigns.Post("/:id/pause", h.Pause)
	campaigns.Post("/:id/resume", h.Resume)
	campaigns.Post("/:id/finish", h.Finish)
	campaigns.Get("/:id/contributions", dashH.Participants)
	campaigns.Get("/:id/contributions/export", dashH.ExportCSV)
	campaigns.Post("/:id/contributions/:contributionId/refund", dashH.Refund)
	campaigns.Post("/:id/images", h.AddImage)
	campaigns.Patch("/:id/images/reorder", h.ReorderImages)
	campaigns.Delete("/:id/images/:imageId", h.DeleteImage)

	router.Post("/files", middleware.RequireAuth(signer), fileH.Upload)
}

func mountPublicRoutes(app fiber.Router, apiV1 fiber.Router, h *handlers.PublicHandler) {
	apiV1.Get("/public/campaigns", h.ListJSON)
	apiV1.Get("/public/campaigns/:slug", h.GetJSON)
	app.Get("/c/:slug", h.OGPage)
}

// mountContributionRoutes implements Etapa 4 §4: el endpoint más importante
// del producto, con rate limit agresivo (10/min por IP).
func mountContributionRoutes(apiV1 fiber.Router, h *handlers.ContributionHandler, contributeLimiter fiber.Handler) {
	apiV1.Post("/public/campaigns/:slug/contributions", contributeLimiter, h.Start)
	apiV1.Get("/public/contributions/:id/status", h.Status)
}

func mountWebhookRoutes(apiV1 fiber.Router, h *handlers.WebhookHandler) {
	apiV1.Post("/webhooks/payments/:provider", h.Payments)
}

// mountWebpayRoutes solo se llama si hay un *webpay.Provider configurado
// (ver server.go) — con el mock, estas rutas ni se registran.
func mountWebpayRoutes(apiV1 fiber.Router, h *handlers.WebpayHandler) {
	apiV1.Get("/webpay/redirect", h.Redirect)
	apiV1.Post("/webpay/return", h.Return)
}

func mountMetaRoutes(router fiber.Router) {
	router.Get("/meta/campaign-types", handlers.CampaignTypes)
}
