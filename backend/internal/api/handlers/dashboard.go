package handlers

import (
	"bytes"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
	dashboarduc "github.com/pcornejov/juntalo/backend/internal/app/dashboard"
)

type DashboardHandler struct {
	participants *dashboarduc.ParticipantsService
	export       *dashboarduc.ExportCSVService
	orgs         app.OrganizationRepository
}

func NewDashboardHandler(participants *dashboarduc.ParticipantsService, export *dashboarduc.ExportCSVService, orgs app.OrganizationRepository) *DashboardHandler {
	return &DashboardHandler{participants: participants, export: export, orgs: orgs}
}

func (h *DashboardHandler) orgID(c *fiber.Ctx) (uuid.UUID, error) {
	org, err := h.orgs.GetPersonalByUserID(c.Context(), middleware.UserID(c))
	if err != nil {
		return uuid.Nil, err
	}
	return org.ID, nil
}

const (
	defaultParticipantsPageSize = 20
	maxParticipantsPageSize     = 100
)

func (h *DashboardHandler) Participants(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	campaignID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return dto.WriteError(c, dashboarduc.ErrCampaignNotFound)
	}

	limit := c.QueryInt("limit", defaultParticipantsPageSize)
	if limit <= 0 || limit > maxParticipantsPageSize {
		limit = defaultParticipantsPageSize
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	rows, err := h.participants.List(c.Context(), campaignID, orgID, int32(limit+1), int32(offset))
	if err != nil {
		return dto.WriteError(c, err)
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	out := make([]dto.ParticipantResponse, len(rows))
	for i, r := range rows {
		out[i] = dto.ParticipantResponse{
			ContributionID: r.ContributionID.String(),
			FullName:       r.FullName,
			Email:          r.Email,
			Phone:          r.Phone,
			Amount:         int64(r.Amount),
			RefundedAmount: int64(r.RefundedAmount),
			IsAnonymous:    r.IsAnonymous,
			Status:         string(r.Status),
			CreatedAt:      r.CreatedAt,
		}
	}
	return c.JSON(dto.ParticipantsResponse{Items: out, HasMore: hasMore})
}

func (h *DashboardHandler) ExportCSV(c *fiber.Ctx) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return dto.WriteError(c, err)
	}
	campaignID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return dto.WriteError(c, dashboarduc.ErrCampaignNotFound)
	}

	var buf bytes.Buffer
	if err := h.export.WriteCSV(c.Context(), campaignID, orgID, &buf); err != nil {
		return dto.WriteError(c, err)
	}

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="participantes.csv"`)
	return c.Send(buf.Bytes())
}
