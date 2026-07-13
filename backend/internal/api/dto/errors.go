package dto

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
)

// ErrorResponse is the single error envelope used across the entire API (Etapa 4 §1).
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// codeStatus maps stable domain error codes to HTTP status codes.
var codeStatus = map[string]int{
	"weak_password":             fiber.StatusUnprocessableEntity,
	"email_already_registered":  fiber.StatusUnprocessableEntity,
	"invalid_credentials":       fiber.StatusUnauthorized,
	"session_expired":           fiber.StatusUnauthorized,
	"unauthorized":              fiber.StatusUnauthorized,
	"validation_failed":         fiber.StatusBadRequest,
	"campaign_not_found":        fiber.StatusNotFound,
	"unknown_campaign_type":     fiber.StatusUnprocessableEntity,
	"campaign_type_disabled":    fiber.StatusUnprocessableEntity,
	"goal_below_raised":         fiber.StatusUnprocessableEntity,
	"invalid_status_transition": fiber.StatusConflict,
	"campaign_not_deletable":    fiber.StatusUnprocessableEntity,
	"campaign_not_active":       fiber.StatusUnprocessableEntity,
	"file_too_large":            fiber.StatusUnprocessableEntity,
	"unsupported_file_type":     fiber.StatusUnprocessableEntity,
}

// WriteError maps err to the stable envelope + HTTP status. Unknown errors never
// leak internal details to the client (Etapa 4: solo codes documentados).
func WriteError(c *fiber.Ctx, err error) error {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		status, ok := codeStatus[appErr.Code]
		if !ok {
			status = fiber.StatusUnprocessableEntity
		}
		return c.Status(status).JSON(ErrorResponse{Error: ErrorBody{Code: appErr.Code, Message: appErr.Message}})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Error: ErrorBody{Code: "internal_error", Message: "Ocurrió un error inesperado"},
	})
}
