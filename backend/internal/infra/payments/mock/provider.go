// Package mock implements app.PaymentProvider without touching a real gateway
// (Etapa 2 §2.3). Modela el ciclo de vida completo — incluyendo confirmación
// asíncrona vía webhook — para que integrar Webpay/Mercado Pago después sea
// implementar la interfaz, no rediseñar el flujo.
package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

// Mode controls how CreateIntent resolves, simulando distintos escenarios de
// una pasarela real.
type Mode string

const (
	// ModeInstant confirma el pago sincrónicamente al crear el intent.
	ModeInstant Mode = "instant"
	// ModeDeferred deja el pago pending y lo confirma poco después vía un
	// webhook real (autoinvocado), ensayando el flujo asíncrono de producción.
	ModeDeferred Mode = "deferred"
	// ModeFail rechaza el pago sincrónicamente.
	ModeFail Mode = "fail"
)

const deferredConfirmDelay = 1500 * time.Millisecond

type Provider struct {
	mode       Mode
	webhookURL string
	secret     string

	mu      sync.Mutex
	intents map[string]string // idempotencyKey -> providerRef, para reintentos/duplicados (Etapa 1 §2)
}

func NewProvider(mode Mode, webhookURL, secret string) *Provider {
	return &Provider{
		mode:       mode,
		webhookURL: webhookURL,
		secret:     secret,
		intents:    map[string]string{},
	}
}

// CreateIntent es idempotente por IdempotencyKey a nivel del propio provider:
// la misma key dos veces devuelve el mismo ProviderRef (Etapa 2 §2.3), como
// haría una pasarela real.
func (p *Provider) CreateIntent(_ context.Context, req app.IntentRequest) (app.PaymentIntent, error) {
	p.mu.Lock()
	if ref, exists := p.intents[req.IdempotencyKey]; exists {
		p.mu.Unlock()
		return app.PaymentIntent{ProviderRef: ref, Status: payment.StatusPending}, nil
	}
	providerRef := "mock_" + uuid.New().String()
	p.intents[req.IdempotencyKey] = providerRef
	p.mu.Unlock()

	switch p.mode {
	case ModeInstant:
		return app.PaymentIntent{ProviderRef: providerRef, Status: payment.StatusConfirmed}, nil
	case ModeFail:
		return app.PaymentIntent{ProviderRef: providerRef, Status: payment.StatusFailed}, nil
	default: // ModeDeferred
		go p.confirmAsync(providerRef)
		return app.PaymentIntent{ProviderRef: providerRef, Status: payment.StatusPending}, nil
	}
}

func (p *Provider) GetIntent(_ context.Context, providerRef string) (app.PaymentIntent, error) {
	// El mock no mantiene estado propio más allá del ref; la fuente de verdad
	// de estado es siempre la tabla payments, actualizada vía webhook.
	return app.PaymentIntent{ProviderRef: providerRef, Status: payment.StatusPending}, nil
}

// Refund simula un reembolso exitoso e instantáneo — no hay dinero real
// involucrado, así que no hay nada que pueda fallar del lado del proveedor.
func (p *Provider) Refund(_ context.Context, providerRef string, amount money.CLP) (app.RefundResult, error) {
	return app.RefundResult{ProviderRef: "mock_refund_" + uuid.New().String(), Amount: amount}, nil
}

// confirmAsync simula la confirmación asíncrona de una pasarela real: espera
// un momento y luego llama al webhook por HTTP, firmado — el mismo camino que
// usará Webpay/Mercado Pago en producción (Etapa 2 §2.3, Etapa 4 §5).
func (p *Provider) confirmAsync(providerRef string) {
	time.Sleep(deferredConfirmDelay)
	sendWebhook(p.webhookURL, p.secret, providerRef, "payment.confirmed")
}
