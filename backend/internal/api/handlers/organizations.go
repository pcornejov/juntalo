package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
	organizationsuc "github.com/pcornejov/juntalo/backend/internal/app/organizations"
)

type OrganizationHandler struct {
	svc  *organizationsuc.Service
	orgs app.OrganizationRepository
}

func NewOrganizationHandler(svc *organizationsuc.Service, orgs app.OrganizationRepository) *OrganizationHandler {
	return &OrganizationHandler{svc: svc, orgs: orgs}
}

// UpdatePayoutInfo implements PATCH /organizations/me/payout: el organizador
// carga sus datos bancarios para recibir la liquidación manual del
// backoffice (Etapa post-MVP: "quién junta el dinero y cuándo transfiere").
func (h *OrganizationHandler) UpdatePayoutInfo(c *fiber.Ctx) error {
	org, err := h.orgs.GetPersonalByUserID(c.Context(), middleware.UserID(c))
	if err != nil {
		return dto.WriteError(c, err)
	}

	var req dto.UpdatePayoutInfoRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	updated, err := h.svc.UpdatePayoutInfo(c.Context(), org.ID, app.UpdatePayoutInfoInput{
		Rut:                 req.Rut,
		PayoutBank:          req.PayoutBank,
		PayoutAccountType:   req.PayoutAccountType,
		PayoutAccountNumber: req.PayoutAccountNumber,
		PayoutHolderName:    req.PayoutHolderName,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}

	return c.JSON(dto.PayoutInfoResponse{
		Rut:                 updated.Rut,
		PayoutBank:          updated.PayoutBank,
		PayoutAccountType:   updated.PayoutAccountType,
		PayoutAccountNumber: updated.PayoutAccountNumber,
		PayoutHolderName:    updated.PayoutHolderName,
	})
}
