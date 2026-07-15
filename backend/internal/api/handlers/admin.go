package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	adminuc "github.com/pcornejov/juntalo/backend/internal/app/admin"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

const (
	defaultAdminListLimit = 20
	maxAdminListLimit     = 100
)

type AdminHandler struct {
	svc *adminuc.Service
}

func NewAdminHandler(svc *adminuc.Service) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) adminListParams(c *fiber.Ctx) (limit, offset int32) {
	l := c.QueryInt("limit", defaultAdminListLimit)
	if l <= 0 || l > maxAdminListLimit {
		l = defaultAdminListLimit
	}
	o := c.QueryInt("offset", 0)
	if o < 0 {
		o = 0
	}
	return int32(l), int32(o)
}

func (h *AdminHandler) Metrics(c *fiber.Ctx) error {
	metrics, err := h.svc.Metrics(c.Context())
	if err != nil {
		return dto.WriteError(c, err)
	}
	return c.JSON(dto.AdminMetricsResponse{
		TotalUsers:         metrics.TotalUsers,
		TotalCampaigns:     metrics.TotalCampaigns,
		ActiveCampaigns:    metrics.ActiveCampaigns,
		DraftCampaigns:     metrics.DraftCampaigns,
		FinishedCampaigns:  metrics.FinishedCampaigns,
		TotalContributions: metrics.TotalContributions,
		RaisedGross:        int64(metrics.RaisedGross),
		RaisedNetApprox:    int64(metrics.RaisedNetApprox),
		TotalCommission:    int64(metrics.TotalCommission),
	})
}

// Users, Campaigns y Payments piden un ítem de más para saber si hay página
// siguiente sin una query COUNT aparte — mismo patrón usado en el resto de
// los listados paginados de la app.
func (h *AdminHandler) Users(c *fiber.Ctx) error {
	limit, offset := h.adminListParams(c)
	items, err := h.svc.ListUsers(c.Context(), limit+1, offset)
	if err != nil {
		return dto.WriteError(c, err)
	}
	hasMore := int32(len(items)) > limit
	if hasMore {
		items = items[:limit]
	}
	out := make([]dto.AdminUserResponse, len(items))
	for i, u := range items {
		out[i] = dto.AdminUserResponse{
			ID:                         u.ID.String(),
			Email:                      u.Email,
			FullName:                   u.FullName,
			EmailVerified:              u.EmailVerified,
			CreatedAt:                  u.CreatedAt,
			OrganizationID:             u.OrganizationID.String(),
			OrganizationName:           u.OrganizationName,
			OrganizationCommissionRate: u.OrganizationCommissionRate,
			CampaignCount:              u.CampaignCount,
		}
	}
	return c.JSON(dto.AdminUserListResponse{Items: out, HasMore: hasMore})
}

// UpdateOrgCommissionRate implements PATCH /admin/organizations/:id/commission.
func (h *AdminHandler) UpdateOrgCommissionRate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return dto.WriteError(c, adminuc.ErrNotFound)
	}
	var req dto.UpdateCommissionRateRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}
	if _, err := h.svc.UpdateOrgCommissionRate(c.Context(), id, req.Rate); err != nil {
		return dto.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) Campaigns(c *fiber.Ctx) error {
	limit, offset := h.adminListParams(c)
	items, err := h.svc.ListCampaigns(c.Context(), limit+1, offset)
	if err != nil {
		return dto.WriteError(c, err)
	}
	hasMore := int32(len(items)) > limit
	if hasMore {
		items = items[:limit]
	}
	out := make([]dto.AdminCampaignResponse, len(items))
	for i, camp := range items {
		out[i] = dto.AdminCampaignResponse{
			ID:               camp.ID.String(),
			Title:            camp.Title,
			Slug:             camp.Slug,
			TypeKey:          string(camp.TypeKey),
			Status:           string(camp.Status),
			Category:         string(camp.Category),
			CreatedAt:        camp.CreatedAt,
			OrganizationName: camp.OrganizationName,
			OrganizerEmail:   camp.OrganizerEmail,
			RaisedGross:      int64(camp.RaisedGross),
			ContributorCount: camp.ContributorCount,
		}
	}
	return c.JSON(dto.AdminCampaignListResponse{Items: out, HasMore: hasMore})
}

// DeleteCampaign implements DELETE /admin/campaigns/:id: la herramienta de
// moderación del backoffice — elimina cualquier campaña de cualquier
// organización, sin restricción de estado (a diferencia del borrado propio
// del organizador, que solo permite borradores).
func (h *AdminHandler) DeleteCampaign(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return dto.WriteError(c, adminuc.ErrNotFound)
	}
	if err := h.svc.DeleteCampaign(c.Context(), id); err != nil {
		return dto.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// PendingPayouts implements GET /admin/payouts/pending: qué organizaciones
// tienen plata confirmada sin liquidar todavía.
func (h *AdminHandler) PendingPayouts(c *fiber.Ctx) error {
	limit, offset := h.adminListParams(c)
	items, err := h.svc.ListPendingPayouts(c.Context(), limit+1, offset)
	if err != nil {
		return dto.WriteError(c, err)
	}
	hasMore := int32(len(items)) > limit
	if hasMore {
		items = items[:limit]
	}
	out := make([]dto.PendingPayoutResponse, len(items))
	for i, p := range items {
		out[i] = dto.PendingPayoutResponse{
			OrganizationID:      p.OrganizationID.String(),
			OrganizationName:    p.OrganizationName,
			Rut:                 p.Rut,
			PayoutBank:          p.PayoutBank,
			PayoutAccountType:   p.PayoutAccountType,
			PayoutAccountNumber: p.PayoutAccountNumber,
			PayoutHolderName:    p.PayoutHolderName,
			EligibleNet:         int64(p.EligibleNet),
			TotalPaid:           int64(p.TotalPaid),
			PendingAmount:       int64(p.PendingAmount),
		}
	}
	return c.JSON(dto.PendingPayoutListResponse{Items: out, HasMore: hasMore})
}

// CreatePayout implements POST /admin/organizations/:id/payouts: deja
// constancia de una transferencia manual que el operador ya hizo desde su
// banco — no dispara ningún movimiento de dinero real.
func (h *AdminHandler) CreatePayout(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return dto.WriteError(c, adminuc.ErrNotFound)
	}
	var req dto.CreatePayoutRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}
	record, err := h.svc.CreatePayout(c.Context(), orgID, money.CLP(req.Amount), req.Note, middleware.UserID(c))
	if err != nil {
		return dto.WriteError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.PayoutResponse{
		ID:             record.ID.String(),
		OrganizationID: record.OrganizationID.String(),
		Amount:         int64(record.Amount),
		Note:           record.Note,
		CreatedAt:      record.CreatedAt,
	})
}

func (h *AdminHandler) Payments(c *fiber.Ctx) error {
	limit, offset := h.adminListParams(c)
	items, err := h.svc.ListPayments(c.Context(), limit+1, offset)
	if err != nil {
		return dto.WriteError(c, err)
	}
	hasMore := int32(len(items)) > limit
	if hasMore {
		items = items[:limit]
	}
	out := make([]dto.AdminPaymentResponse, len(items))
	for i, p := range items {
		out[i] = dto.AdminPaymentResponse{
			ID:               p.ID.String(),
			Status:           string(p.Status),
			Provider:         p.Provider,
			AmountGross:      int64(p.AmountGross),
			AmountNet:        int64(p.AmountNet),
			CommissionAmount: int64(p.CommissionAmount),
			CreatedAt:        p.CreatedAt,
			ConfirmedAt:      p.ConfirmedAt,
			CampaignTitle:    p.CampaignTitle,
			CampaignSlug:     p.CampaignSlug,
			ContributorName:  p.ContributorName,
		}
	}
	return c.JSON(dto.AdminPaymentListResponse{Items: out, HasMore: hasMore})
}
