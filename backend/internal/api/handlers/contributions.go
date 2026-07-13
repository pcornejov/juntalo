package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	contributionsuc "github.com/pcornejov/juntalo/backend/internal/app/contributions"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

var errMissingIdempotencyKey = apperr.New("validation_failed", "Falta el header Idempotency-Key")

type ContributionHandler struct {
	start  *contributionsuc.StartService
	status *contributionsuc.StatusService
}

func NewContributionHandler(start *contributionsuc.StartService, status *contributionsuc.StatusService) *ContributionHandler {
	return &ContributionHandler{start: start, status: status}
}

// Start implements el endpoint más importante del producto (Etapa 4 §4).
func (h *ContributionHandler) Start(c *fiber.Ctx) error {
	idempotencyKey := c.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return dto.WriteError(c, errMissingIdempotencyKey)
	}

	var req dto.StartContributionRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	result, err := h.start.Start(c.Context(), contributionsuc.StartInput{
		Slug:           c.Params("slug"),
		IdempotencyKey: idempotencyKey,
		FullName:       req.FullName,
		Email:          req.Email,
		Phone:          req.Phone,
		Amount:         money.CLP(req.Amount),
		IsAnonymous:    req.IsAnonymous,
		Message:        req.Message,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}

	resp := dto.StartContributionResponse{ContributionID: result.ContributionID.String()}
	resp.Payment.Status = string(result.PaymentStatus)
	resp.Payment.RedirectURL = result.RedirectURL
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// Status is polled by the confirmation screen (Etapa 4 §4).
func (h *ContributionHandler) Status(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return dto.WriteError(c, contributionsuc.ErrNotFound)
	}

	status, err := h.status.GetStatus(c.Context(), id)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return c.JSON(dto.ContributionStatusResponse{Status: string(status)})
}
