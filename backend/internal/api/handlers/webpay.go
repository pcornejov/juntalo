package handlers

import (
	"fmt"
	"html"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	contributionsuc "github.com/pcornejov/juntalo/backend/internal/app/contributions"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
	"github.com/pcornejov/juntalo/backend/internal/infra/payments/webpay"
)

// WebpayHandler implementa las dos rutas que Webpay Plus necesita fuera del
// ciclo normal request/response de una API: un helper de redirect (Transbank
// exige un form POST con el token, no un simple GET) y el return_url al que
// el navegador del aportante vuelve después de pagar.
type WebpayHandler struct {
	provider      *webpay.Provider
	confirm       *contributionsuc.ConfirmService
	contributions app.ContributionRepository
	campaigns     app.CampaignRepository
	frontendURL   string
}

func NewWebpayHandler(
	provider *webpay.Provider,
	confirm *contributionsuc.ConfirmService,
	contributions app.ContributionRepository,
	campaigns app.CampaignRepository,
	frontendURL string,
) *WebpayHandler {
	return &WebpayHandler{
		provider: provider, confirm: confirm,
		contributions: contributions, campaigns: campaigns, frontendURL: frontendURL,
	}
}

// Redirect sirve una página mínima que auto-envía un form POST con el token
// a la URL de Transbank — el paso intermedio que el flujo de Webpay Plus
// exige entre "crear la transacción" y "el aportante entra su tarjeta"
// (Etapa 2 §2.3).
func (h *WebpayHandler) Redirect(c *fiber.Ctx) error {
	token := c.Query("token")
	initURL := c.Query("init_url")
	if token == "" || initURL == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Falta token o init_url")
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(fmt.Sprintf(`<!doctype html>
<html lang="es"><head><meta charset="utf-8"><title>Redirigiendo a Webpay…</title></head>
<body onload="document.forms[0].submit()">
<p>Redirigiendo a Webpay, un momento…</p>
<form method="POST" action="%s">
<input type="hidden" name="token_ws" value="%s">
<noscript><button type="submit">Continuar</button></noscript>
</form>
</body></html>`, html.EscapeString(initURL), html.EscapeString(token)))
}

// Return implementa el return_url que Transbank llama (POST, form-encoded)
// cuando el aportante termina en la página de pago — con éxito, rechazo, o
// abandono. Confirma la transacción, actualiza el pago, y manda al navegador
// de vuelta a la página pública de la campaña para que retome el polling de
// estado normal (Etapa 4 §5).
func (h *WebpayHandler) Return(c *fiber.Ctx) error {
	token := c.FormValue("token_ws")
	aborted := token == "" && c.FormValue("TBK_TOKEN") != ""

	var event string
	if aborted {
		// El aportante anuló la compra en la página de Transbank antes de
		// terminar — no hay nada que confirmar, el pago queda en failed.
		token = c.FormValue("TBK_TOKEN")
		event = "payment.failed"
	} else if token == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Falta token_ws")
	} else {
		status, err := h.provider.Commit(c.Context(), token)
		if err != nil {
			status = payment.StatusFailed
		}
		if status == payment.StatusConfirmed {
			event = "payment.confirmed"
		} else {
			event = "payment.failed"
		}
	}

	confirmedPayment, err := h.confirm.HandleWebhookEvent(c.Context(), "webpay", token, event)
	if err != nil || confirmedPayment.ContributionID == uuid.Nil {
		// Token desconocido o error interno: no hay contribution_id para armar
		// un redirect con contexto — manda al home del frontend en vez de
		// romper con un error crudo en el navegador del aportante.
		return c.Redirect(h.frontendURL, fiber.StatusFound)
	}

	contrib, found, err := h.contributions.GetByID(c.Context(), confirmedPayment.ContributionID)
	if err != nil || !found {
		return c.Redirect(h.frontendURL, fiber.StatusFound)
	}
	camp, found, err := h.campaigns.GetByID(c.Context(), contrib.CampaignID)
	if err != nil || !found {
		return c.Redirect(h.frontendURL, fiber.StatusFound)
	}

	redirectURL := fmt.Sprintf("%s/public/%s?contribution_id=%s",
		h.frontendURL, camp.Slug, url.QueryEscape(confirmedPayment.ContributionID.String()))
	return c.Redirect(redirectURL, fiber.StatusFound)
}
