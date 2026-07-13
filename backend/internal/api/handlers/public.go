package handlers

import (
	"fmt"
	"html"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/app"
	campaignsuc "github.com/pcornejov/juntalo/backend/internal/app/campaigns"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

// botUserAgents identifies link-preview crawlers that need server-rendered OG
// tags without executing JS (Etapa 2 §1, Etapa 4 §4).
var botUserAgents = []string{
	"whatsapp", "facebookexternalhit", "twitterbot", "slackbot",
	"linkedinbot", "telegrambot", "discordbot", "googlebot",
}

type PublicHandler struct {
	get         *campaignsuc.GetService
	files       app.FileRepository
	storage     app.FileStorage
	frontendURL string
}

func NewPublicHandler(get *campaignsuc.GetService, files app.FileRepository, storage app.FileStorage, frontendURL string) *PublicHandler {
	return &PublicHandler{get: get, files: files, storage: storage, frontendURL: frontendURL}
}

// GetJSON is consumed by the SPA's public campaign page (Etapa 4 §4).
func (h *PublicHandler) GetJSON(c *fiber.Ctx) error {
	c.Set("Cache-Control", "public, max-age=30")

	found, totals, err := h.get.GetPublicBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return dto.WriteError(c, err)
	}

	coverURL := resolveCoverURL(c, h.files, h.storage, found.CoverFileID)
	return c.JSON(toPublicCampaignResponse(found, totals, coverURL))
}

// OGPage serves GET /c/:slug: HTML con OG tags para bots/previews, 302 a la
// SPA para navegadores reales (Etapa 2 §1, Etapa 4 §4).
func (h *PublicHandler) OGPage(c *fiber.Ctx) error {
	found, totals, err := h.get.GetPublicBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Campaña no encontrada")
	}

	redirectPath := fmt.Sprintf("%s/public/%s", h.frontendURL, found.Slug)

	if !isBotRequest(c.Get("User-Agent")) {
		return c.Redirect(redirectPath, fiber.StatusFound)
	}

	coverURL := resolveCoverURL(c, h.files, h.storage, found.CoverFileID)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(renderOGHTML(found, totals, coverURL, redirectPath))
}

func isBotRequest(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	for _, b := range botUserAgents {
		if strings.Contains(ua, b) {
			return true
		}
	}
	return false
}

func resolveCoverURL(c *fiber.Ctx, files app.FileRepository, storage app.FileStorage, coverFileID *uuid.UUID) *string {
	if coverFileID == nil {
		return nil
	}
	file, found, err := files.GetByID(c.Context(), *coverFileID)
	if err != nil || !found {
		return nil
	}
	url := storage.PublicURL(file.StorageKey)
	return &url
}

func toPublicCampaignResponse(c campaign.Campaign, totals campaign.Totals, coverURL *string) dto.PublicCampaignResponse {
	def := campaign.Registry[c.TypeKey]
	resp := dto.PublicCampaignResponse{
		Title:       c.Title,
		Description: c.Description,
		CoverURL:    coverURL,
		Status:      string(c.Status),
		CTA:         def.Labels.CTA,
		Unit:        def.Labels.Unit,
		Totals: dto.TotalsDTO{
			RaisedGross:      int64(totals.RaisedGross),
			RaisedNetApprox:  int64(totals.RaisedNetApprox),
			ContributorCount: totals.ContributorCount,
		},
	}
	if c.GoalAmount != nil {
		v := int64(*c.GoalAmount)
		resp.GoalAmount = &v
	}
	return resp
}

func renderOGHTML(c campaign.Campaign, totals campaign.Totals, coverURL *string, redirectPath string) string {
	title := html.EscapeString(c.Title)
	description := html.EscapeString(truncate(c.Description, 200))

	image := ""
	if coverURL != nil {
		image = fmt.Sprintf(`<meta property="og:image" content="%s">`, html.EscapeString(*coverURL))
	}

	return fmt.Sprintf(`<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>%s — Juntalo</title>
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:type" content="website">
%s
<meta http-equiv="refresh" content="0; url=%s">
</head>
<body>
<p>%s</p>
<p><a href="%s">Ver campaña</a></p>
</body>
</html>`, title, title, description, image, html.EscapeString(redirectPath), description, html.EscapeString(redirectPath))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
