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
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

// errOrgNotFound: un slug de organización inexistente responde 404 igual que
// una campaña ajena — no confirma ni niega su existencia.
var errOrgNotFound = apperr.New("organization_not_found", "Organización no encontrada")

// botUserAgents identifies link-preview crawlers that need server-rendered OG
// tags without executing JS (Etapa 2 §1, Etapa 4 §4).
var botUserAgents = []string{
	"whatsapp", "facebookexternalhit", "twitterbot", "slackbot",
	"linkedinbot", "telegrambot", "discordbot", "googlebot",
}

type PublicHandler struct {
	get         *campaignsuc.GetService
	list        *campaignsuc.ListService
	files       app.FileRepository
	images      app.CampaignImageRepository
	storage     app.FileStorage
	orgs        app.OrganizationRepository
	frontendURL string
	selfURL     string
}

func NewPublicHandler(get *campaignsuc.GetService, list *campaignsuc.ListService, files app.FileRepository, images app.CampaignImageRepository, storage app.FileStorage, orgs app.OrganizationRepository, frontendURL, selfURL string) *PublicHandler {
	return &PublicHandler{get: get, list: list, files: files, images: images, storage: storage, orgs: orgs, frontendURL: frontendURL, selfURL: selfURL}
}

const (
	defaultPublicListLimit = 12
	maxPublicListLimit     = 50
)

// ListJSON implements GET /public/campaigns: la sección pública de
// "Campañas activas" (Etapa 4) — cualquier visitante puede explorar
// campañas para aportar, no solo entrar por un link directo.
func (h *PublicHandler) ListJSON(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", defaultPublicListLimit)
	if limit <= 0 || limit > maxPublicListLimit {
		limit = defaultPublicListLimit
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}
	search := c.Query("q")
	category := c.Query("category")

	items, err := h.list.ListPublic(c.Context(), search, campaign.Category(category), int32(limit+1), int32(offset))
	if err != nil {
		return dto.WriteError(c, err)
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	out := make([]dto.PublicCampaignResponse, len(items))
	for i, item := range items {
		images := resolveGalleryURLs(c.Context(), h.images, h.storage, item.Campaign.ID)
		coverURL := resolveCoverURL(c, h.files, h.storage, item.Campaign.CoverFileID)
		if len(images) > 0 {
			coverURL = &images[0]
		}
		organizerName, isVerified := h.resolveOrganizer(c, item.Campaign.OrganizationID)
		resp := toPublicCampaignResponse(item.Campaign, item.Totals, coverURL, images, item.RaffleNumbersSold)
		resp.PublicURL = h.selfURL + "/c/" + item.Campaign.Slug
		resp.OrganizerName = organizerName
		resp.IsVerified = isVerified
		out[i] = resp
	}
	return c.JSON(dto.PublicCampaignListResponse{Items: out, HasMore: hasMore})
}

// Featured implements GET /public/campaigns/featured: la campaña "más
// caliente" (inspirado en Vaki) para destacar en la sección pública.
func (h *PublicHandler) Featured(c *fiber.Ctx) error {
	item, found, err := h.list.GetFeatured(c.Context())
	if err != nil {
		return dto.WriteError(c, err)
	}
	if !found {
		return c.Status(fiber.StatusNoContent).Send(nil)
	}

	images := resolveGalleryURLs(c.Context(), h.images, h.storage, item.Campaign.ID)
	coverURL := resolveCoverURL(c, h.files, h.storage, item.Campaign.CoverFileID)
	if len(images) > 0 {
		coverURL = &images[0]
	}
	organizerName, isVerified := h.resolveOrganizer(c, item.Campaign.OrganizationID)
	resp := toPublicCampaignResponse(item.Campaign, item.Totals, coverURL, images, item.RaffleNumbersSold)
	resp.PublicURL = h.selfURL + "/c/" + item.Campaign.Slug
	resp.OrganizerName = organizerName
	resp.IsVerified = isVerified
	return c.JSON(resp)
}

// resolveOrganizer no falla el request si la búsqueda del dueño falla — el
// badge de verificación es puramente decorativo, no debe tumbar la página
// pública.
func (h *PublicHandler) resolveOrganizer(c *fiber.Ctx, orgID uuid.UUID) (name string, verified bool) {
	name, verified, err := h.orgs.GetOwnerInfo(c.Context(), orgID)
	if err != nil {
		return "", false
	}
	return name, verified
}

// OrgProfile implements GET /public/organizations/:slug: la página pública
// persistente del organizador (inspirada en el link único de por vida de
// Ceneka) — a diferencia de una campaña puntual, lista todas las campañas
// visibles (active/paused/finished) de esa organización.
func (h *PublicHandler) OrgProfile(c *fiber.Ctx) error {
	slug := c.Params("slug")
	org, found, err := h.orgs.GetBySlug(c.Context(), slug)
	if err != nil {
		return dto.WriteError(c, err)
	}
	if !found {
		return dto.WriteError(c, errOrgNotFound)
	}

	limit := c.QueryInt("limit", defaultPublicListLimit)
	if limit <= 0 || limit > maxPublicListLimit {
		limit = defaultPublicListLimit
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	items, err := h.list.ListPublicByOrg(c.Context(), org.ID, int32(limit+1), int32(offset))
	if err != nil {
		return dto.WriteError(c, err)
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	_, isVerified := h.resolveOrganizer(c, org.ID)

	campaigns := make([]dto.PublicCampaignResponse, len(items))
	for i, item := range items {
		images := resolveGalleryURLs(c.Context(), h.images, h.storage, item.Campaign.ID)
		coverURL := resolveCoverURL(c, h.files, h.storage, item.Campaign.CoverFileID)
		if len(images) > 0 {
			coverURL = &images[0]
		}
		resp := toPublicCampaignResponse(item.Campaign, item.Totals, coverURL, images, item.RaffleNumbersSold)
		resp.PublicURL = h.selfURL + "/c/" + item.Campaign.Slug
		resp.OrganizerName = org.Name
		resp.IsVerified = isVerified
		campaigns[i] = resp
	}

	return c.JSON(dto.OrgProfileResponse{
		Name:       org.Name,
		Slug:       org.Slug,
		IsVerified: isVerified,
		Campaigns:  campaigns,
		HasMore:    hasMore,
	})
}

// GetJSON is consumed by the SPA's public campaign page (Etapa 4 §4).
func (h *PublicHandler) GetJSON(c *fiber.Ctx) error {
	// Antes tenía Cache-Control: public, max-age=30 — pero el navegador
	// respeta ese header a nivel de fetch() sin importar que React Query
	// invalide su propia caché, así que el refetch tras confirmar un pago
	// podía seguir devolviendo el total viejo hasta por 30s. El total y el
	// contador de aportantes son justamente lo que más necesita ser
	// correcto en tiempo real, así que no vale la pena cachear esta
	// respuesta a nivel HTTP.
	c.Set("Cache-Control", "no-store")

	slug := c.Params("slug")
	found, totals, err := h.get.GetPublicBySlug(c.Context(), slug)
	if err != nil {
		return dto.WriteError(c, err)
	}

	images := resolveGalleryURLs(c.Context(), h.images, h.storage, found.ID)
	coverURL := resolveCoverURL(c, h.files, h.storage, found.CoverFileID)
	if len(images) > 0 {
		coverURL = &images[0]
	}
	var raffleSold *int64
	if found.TypeKey == campaign.TypeRaffle {
		if sold, err := h.get.GetRaffleNumbersSold(c.Context(), found.ID); err == nil {
			raffleSold = &sold
		}
	}
	resp := toPublicCampaignResponse(found, totals, coverURL, images, raffleSold)
	// public_url apunta a /c/:slug en el dominio del backend (donde vive el
	// render de OG tags), no al dominio del frontend — así, si alguien
	// re-comparte el link desde la página pública, la preview de WhatsApp
	// sigue funcionando (Etapa 2 §1).
	resp.PublicURL = h.selfURL + "/c/" + slug
	resp.OrganizerName, resp.IsVerified = h.resolveOrganizer(c, found.OrganizationID)
	return c.JSON(resp)
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

	images := resolveGalleryURLs(c.Context(), h.images, h.storage, found.ID)
	coverURL := resolveCoverURL(c, h.files, h.storage, found.CoverFileID)
	if len(images) > 0 {
		coverURL = &images[0]
	}
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

func toPublicCampaignResponse(c campaign.Campaign, totals campaign.Totals, coverURL *string, images []string, raffleSold *int64) dto.PublicCampaignResponse {
	def := campaign.Registry[c.TypeKey]
	resp := dto.PublicCampaignResponse{
		Slug:        c.Slug,
		Title:       c.Title,
		Description: c.Description,
		CoverURL:    coverURL,
		Images:      images,
		Status:      string(c.Status),
		Category:    string(c.Category),
		VideoURL:    c.VideoURL,
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
	if c.RaffleUnitPrice != nil {
		v := int64(*c.RaffleUnitPrice)
		resp.RaffleUnitPrice = &v
	}
	resp.RaffleTotalNumbers = c.RaffleTotalNumbers
	resp.RaffleWinningNumber = c.RaffleWinningNumber
	if c.RaffleTotalNumbers != nil && raffleSold != nil {
		available := int64(*c.RaffleTotalNumbers) - *raffleSold
		if available < 0 {
			available = 0
		}
		resp.RaffleAvailable = &available
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
