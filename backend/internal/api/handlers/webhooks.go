package handlers

import (
	"crypto/hmac"

	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	contributionsuc "github.com/pcornejov/juntalo/backend/internal/app/contributions"
	"github.com/pcornejov/juntalo/backend/internal/infra/payments/mock"
)

type WebhookHandler struct {
	confirm    *contributionsuc.ConfirmService
	mockSecret string
}

func NewWebhookHandler(confirm *contributionsuc.ConfirmService, mockSecret string) *WebhookHandler {
	return &WebhookHandler{confirm: confirm, mockSecret: mockSecret}
}

// Payments implements POST /webhooks/payments/:provider (Etapa 4 §5). Valida
// la firma; eventos con firma inválida se rechazan, eventos desconocidos con
// firma válida se aceptan como no-op (200) para no amplificar reintentos del
// proveedor. Idempotente por construcción (ConfirmByProviderRef en dominio).
//
// El :provider en la URL es allowlist explícita (default-deny), no un "si
// coincide con 'mock' valido, si no dejo pasar" — auditoría de seguridad:
// la versión anterior aceptaba sin validar firma cualquier :provider que no
// fuera literalmente "mock", un patrón frágil para cuando se agregue un
// proveedor real y alguien olvide replicar la verificación.
func (h *WebhookHandler) Payments(c *fiber.Ctx) error {
	provider := c.Params("provider")
	body := c.Body()

	switch provider {
	case "mock":
		expected := mock.Sign(h.mockSecret, body)
		if !hmac.Equal([]byte(c.Get("X-Signature")), []byte(expected)) {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
	default:
		return c.SendStatus(fiber.StatusNotFound)
	}

	var payload dto.WebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return dto.WriteError(c, err)
	}

	if err := h.confirm.HandleWebhookEvent(c.Context(), provider, payload.ProviderRef, payload.Event); err != nil {
		// Referencia desconocida: puede ser ruido del proveedor (reintento
		// tardío de un test, etc.) — se registra pero no se amplifica con 4xx.
		if contributionsuc.IsPaymentNotFound(err) {
			return c.SendStatus(fiber.StatusOK)
		}
		return dto.WriteError(c, err)
	}

	return c.SendStatus(fiber.StatusOK)
}
