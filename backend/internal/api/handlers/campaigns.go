package handlers

import (
	"context"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
	campaignsuc "github.com/pcornejov/juntalo/backend/internal/app/campaigns"
	filesuc "github.com/pcornejov/juntalo/backend/internal/app/files"
	auditcat "github.com/pcornejov/juntalo/backend/internal/domain/audit"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

const galleryUploadKind = "campaign_gallery"

const defaultListLimit = 20
const maxListLimit = 50

type CampaignHandler struct {
	create     *campaignsuc.CreateService
	get        *campaignsuc.GetService
	list       *campaignsuc.ListService
	update     *campaignsuc.UpdateService
	transition *campaignsuc.TransitionService
	del        *campaignsuc.DeleteService
	clone      *campaignsuc.CloneService
	upload     *filesuc.UploadService
	orgs       app.OrganizationRepository
	files      app.FileRepository
	images     app.CampaignImageRepository
	storage    app.FileStorage
	audit      app.AuditRepository
	selfURL    string
}

func NewCampaignHandler(
	create *campaignsuc.CreateService,
	get *campaignsuc.GetService,
	list *campaignsuc.ListService,
	update *campaignsuc.UpdateService,
	transition *campaignsuc.TransitionService,
	del *campaignsuc.DeleteService,
	clone *campaignsuc.CloneService,
	upload *filesuc.UploadService,
	orgs app.OrganizationRepository,
	files app.FileRepository,
	images app.CampaignImageRepository,
	storage app.FileStorage,
	audit app.AuditRepository,
	selfURL string,
) *CampaignHandler {
	return &CampaignHandler{
		create: create, get: get, list: list, update: update,
		transition: transition, del: del, clone: clone, upload: upload, orgs: orgs,
		files: files, images: images, storage: storage, audit: audit, selfURL: selfURL,
	}
}

// recordAudit logs a campaign lifecycle event without failing the request if
// the write itself has a problem — auditoría no debe poder tumbar el flujo
// principal del organizador.
func (h *CampaignHandler) recordAudit(c *fiber.Ctx, orgID uuid.UUID, action string, campaignID uuid.UUID, data map[string]any) {
	_ = h.audit.Record(c.Context(), app.RecordAuditInput{
		ActorUserID:    middleware.UserID(c),
		OrganizationID: orgID,
		Action:         action,
		EntityType:     auditcat.EntityTypeCampaign,
		EntityID:       campaignID,
		Data:           data,
	})
}

func (h *CampaignHandler) orgID(c *fiber.Ctx) (uuid.UUID, error) {
	userID := middleware.UserID(c)
	org, err := h.orgs.GetPersonalByUserID(c.Context(), userID)
	if err != nil {
		return uuid.Nil, err
	}
	return org.ID, nil
}

func (h *CampaignHandler) parseID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, campaignsuc.ErrNotFound
	}
	return id, nil
}

func (h *CampaignHandler) Create(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	var req dto.CreateCampaignRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	created, err := h.create.Create(c.Context(), campaignsuc.CreateInput{
		OrganizationID: orgID,
		TypeKey:        campaign.TypeKey(req.TypeKey),
		Title:          req.Title,
		Description:    req.Description,
		GoalAmount:     goalFromRequest(req.GoalAmount),
		StartsAt:       req.StartsAt,
		EndsAt:         req.EndsAt,
		PublishAt:      req.PublishAt,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}
	h.recordAudit(c, orgID, auditcat.ActionCampaignCreated, created.ID, map[string]any{"title": created.Title})

	return c.Status(fiber.StatusCreated).JSON(h.toResponse(c, created, campaign.Totals{}))
}

func (h *CampaignHandler) CancelSchedule(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	updated, err := h.update.CancelSchedule(c.Context(), id, orgID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	_, totals, err := h.get.GetForOrg(c.Context(), id, orgID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return c.JSON(h.toResponse(c, updated, totals))
}

func (h *CampaignHandler) Clone(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	cloned, err := h.clone.Clone(c.Context(), id, orgID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	h.recordAudit(c, orgID, auditcat.ActionCampaignCreated, cloned.ID, map[string]any{"title": cloned.Title, "cloned_from": id.String()})

	return c.Status(fiber.StatusCreated).JSON(h.toResponse(c, cloned, campaign.Totals{}))
}

func (h *CampaignHandler) List(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	limit := c.QueryInt("limit", defaultListLimit)
	if limit <= 0 || limit > maxListLimit {
		limit = defaultListLimit
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	// Se pide un ítem de más para saber si hay página siguiente sin una
	// query COUNT aparte — a esta escala (un organizador, sus campañas) es
	// más simple que sumar paginación por cursor con total.
	items, err := h.list.List(c.Context(), orgID, int32(limit+1), int32(offset))
	if err != nil {
		return dto.WriteError(c, err)
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	out := make([]dto.CampaignResponse, len(items))
	for i, item := range items {
		out[i] = h.toResponse(c, item.Campaign, item.Totals)
	}
	return c.JSON(dto.CampaignListResponse{Items: out, HasMore: hasMore})
}

func (h *CampaignHandler) Get(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	found, totals, err := h.get.GetForOrg(c.Context(), id, orgID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return c.JSON(h.toResponse(c, found, totals))
}

func (h *CampaignHandler) Update(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	var req dto.UpdateCampaignRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	var coverFileID *uuid.UUID
	if req.CoverFileID != nil {
		parsed, err := uuid.Parse(*req.CoverFileID)
		if err != nil {
			return dto.WriteError(c, campaignsuc.ErrNotFound)
		}
		coverFileID = &parsed
	}

	if _, err := h.update.Update(c.Context(), id, orgID, campaignsuc.UpdateInput{
		Title:          req.Title,
		Description:    req.Description,
		GoalAmount:     goalFromRequest(req.GoalAmount),
		StartsAt:       req.StartsAt,
		EndsAt:         req.EndsAt,
		CoverFileID:    coverFileID,
		PublishAt:      req.PublishAt,
		ClearPublishAt: req.ClearPublishAt,
	}); err != nil {
		return dto.WriteError(c, err)
	}

	updated, totals, err := h.get.GetForOrg(c.Context(), id, orgID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return c.JSON(h.toResponse(c, updated, totals))
}

func (h *CampaignHandler) Publish(c *fiber.Ctx) error {
	return h.doTransition(c, h.transition.Publish, auditcat.ActionCampaignPublished)
}

func (h *CampaignHandler) Pause(c *fiber.Ctx) error {
	return h.doTransition(c, h.transition.Pause, auditcat.ActionCampaignPaused)
}

func (h *CampaignHandler) Resume(c *fiber.Ctx) error {
	return h.doTransition(c, h.transition.Resume, auditcat.ActionCampaignResumed)
}

func (h *CampaignHandler) Finish(c *fiber.Ctx) error {
	return h.doTransition(c, h.transition.Finish, auditcat.ActionCampaignFinished)
}

func (h *CampaignHandler) doTransition(c *fiber.Ctx, action func(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, error), auditAction string) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	updated, err := action(c.Context(), id, orgID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	h.recordAudit(c, orgID, auditAction, id, nil)

	_, totals, err := h.get.GetForOrg(c.Context(), id, orgID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return c.JSON(h.toResponse(c, updated, totals))
}

func (h *CampaignHandler) Delete(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	if err := h.del.Delete(c.Context(), id, orgID); err != nil {
		return dto.WriteError(c, err)
	}
	h.recordAudit(c, orgID, auditcat.ActionCampaignDeleted, id, nil)
	return c.SendStatus(fiber.StatusNoContent)
}

// AddImage implements el carrusel de fotos: sube directo al storage y la
// asocia a la campaña en un solo paso (a diferencia del flujo viejo de
// cover_file_id, que subía a POST /files y recién después hacía PATCH).
func (h *CampaignHandler) AddImage(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	if _, _, err := h.get.GetForOrg(c.Context(), id, orgID); err != nil {
		return dto.WriteError(c, err)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return dto.WriteError(c, err)
	}
	f, err := fileHeader.Open()
	if err != nil {
		return dto.WriteError(c, err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		return dto.WriteError(c, err)
	}

	uploaded, err := h.upload.Upload(c.Context(), filesuc.UploadInput{
		OrganizationID: orgID,
		Kind:           galleryUploadKind,
		MimeType:       fileHeader.Header.Get("Content-Type"),
		Data:           data,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}

	imageID, err := h.images.Add(c.Context(), id, uploaded.ID)
	if err != nil {
		return dto.WriteError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.CampaignImageResponse{ID: imageID.String(), URL: uploaded.URL})
}

func (h *CampaignHandler) DeleteImage(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	imageID, err := uuid.Parse(c.Params("imageId"))
	if err != nil {
		return dto.WriteError(c, campaignsuc.ErrNotFound)
	}
	if _, _, err := h.get.GetForOrg(c.Context(), id, orgID); err != nil {
		return dto.WriteError(c, err)
	}

	deleted, err := h.images.Delete(c.Context(), id, imageID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	if !deleted {
		return dto.WriteError(c, campaignsuc.ErrNotFound)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ReorderImages recibe el orden completo de la galería; la primera imagen
// queda como portada — "elegir portada" es simplemente moverla al inicio.
func (h *CampaignHandler) ReorderImages(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	id, err := h.parseID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	if _, _, err := h.get.GetForOrg(c.Context(), id, orgID); err != nil {
		return dto.WriteError(c, err)
	}

	var req dto.ReorderImagesRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	imageIDs := make([]uuid.UUID, len(req.ImageIDs))
	for i, raw := range req.ImageIDs {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return dto.WriteError(c, campaignsuc.ErrNotFound)
		}
		imageIDs[i] = parsed
	}

	if err := h.images.Reorder(c.Context(), id, imageIDs); err != nil {
		return dto.WriteError(c, err)
	}

	updated, totals, err := h.get.GetForOrg(c.Context(), id, orgID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return c.JSON(h.toResponse(c, updated, totals))
}

func goalFromRequest(v *int64) *money.CLP {
	if v == nil {
		return nil
	}
	amount := money.CLP(*v)
	return &amount
}

func (h *CampaignHandler) toResponse(c *fiber.Ctx, camp campaign.Campaign, totals campaign.Totals) dto.CampaignResponse {
	images := resolveGalleryImages(c.Context(), h.images, h.storage, camp.ID)
	coverURL := resolveCoverURL(c, h.files, h.storage, camp.CoverFileID)
	if len(images) > 0 {
		coverURL = &images[0].URL
	}
	resp := dto.CampaignResponse{
		ID:          camp.ID.String(),
		TypeKey:     string(camp.TypeKey),
		Title:       camp.Title,
		Slug:        camp.Slug,
		Description: camp.Description,
		CoverURL:    coverURL,
		Images:      images,
		Status:      string(camp.Status),
		StartsAt:    camp.StartsAt,
		EndsAt:      camp.EndsAt,
		PublishAt:   camp.PublishAt,
		PublicURL:   h.selfURL + "/c/" + camp.Slug,
		CreatedAt:   camp.CreatedAt,
		Totals: dto.TotalsDTO{
			RaisedGross:      int64(totals.RaisedGross),
			RaisedNetApprox:  int64(totals.RaisedNetApprox),
			ContributorCount: totals.ContributorCount,
		},
	}
	if camp.GoalAmount != nil {
		v := int64(*camp.GoalAmount)
		resp.GoalAmount = &v
	}
	return resp
}
