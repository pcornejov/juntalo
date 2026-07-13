package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
	campaignsuc "github.com/pcornejov/juntalo/backend/internal/app/campaigns"
	filesuc "github.com/pcornejov/juntalo/backend/internal/app/files"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

const defaultListLimit = 20

type CampaignHandler struct {
	create     *campaignsuc.CreateService
	get        *campaignsuc.GetService
	list       *campaignsuc.ListService
	update     *campaignsuc.UpdateService
	transition *campaignsuc.TransitionService
	del        *campaignsuc.DeleteService
	upload     *filesuc.UploadService
	orgs       app.OrganizationRepository
	files      app.FileRepository
	storage    app.FileStorage
}

func NewCampaignHandler(
	create *campaignsuc.CreateService,
	get *campaignsuc.GetService,
	list *campaignsuc.ListService,
	update *campaignsuc.UpdateService,
	transition *campaignsuc.TransitionService,
	del *campaignsuc.DeleteService,
	upload *filesuc.UploadService,
	orgs app.OrganizationRepository,
	files app.FileRepository,
	storage app.FileStorage,
) *CampaignHandler {
	return &CampaignHandler{
		create: create, get: get, list: list, update: update,
		transition: transition, del: del, upload: upload, orgs: orgs,
		files: files, storage: storage,
	}
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
	})
	if err != nil {
		return dto.WriteError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(h.toResponse(c, created, campaign.Totals{}))
}

func (h *CampaignHandler) List(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}

	items, err := h.list.List(c.Context(), orgID, defaultListLimit, 0)
	if err != nil {
		return dto.WriteError(c, err)
	}

	out := make([]dto.CampaignResponse, len(items))
	for i, item := range items {
		out[i] = h.toResponse(c, item.Campaign, item.Totals)
	}
	return c.JSON(dto.CampaignListResponse{Items: out})
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
		Title:       req.Title,
		Description: req.Description,
		GoalAmount:  goalFromRequest(req.GoalAmount),
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		CoverFileID: coverFileID,
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
	return h.doTransition(c, h.transition.Publish)
}

func (h *CampaignHandler) Pause(c *fiber.Ctx) error {
	return h.doTransition(c, h.transition.Pause)
}

func (h *CampaignHandler) Resume(c *fiber.Ctx) error {
	return h.doTransition(c, h.transition.Resume)
}

func (h *CampaignHandler) Finish(c *fiber.Ctx) error {
	return h.doTransition(c, h.transition.Finish)
}

func (h *CampaignHandler) doTransition(c *fiber.Ctx, action func(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, error)) error {
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
	return c.SendStatus(fiber.StatusNoContent)
}

func goalFromRequest(v *int64) *money.CLP {
	if v == nil {
		return nil
	}
	amount := money.CLP(*v)
	return &amount
}

func (h *CampaignHandler) toResponse(c *fiber.Ctx, camp campaign.Campaign, totals campaign.Totals) dto.CampaignResponse {
	resp := dto.CampaignResponse{
		ID:          camp.ID.String(),
		TypeKey:     string(camp.TypeKey),
		Title:       camp.Title,
		Slug:        camp.Slug,
		Description: camp.Description,
		CoverURL:    resolveCoverURL(c, h.files, h.storage, camp.CoverFileID),
		Status:      string(camp.Status),
		StartsAt:    camp.StartsAt,
		EndsAt:      camp.EndsAt,
		PublicURL:   "/c/" + camp.Slug,
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
