// Package webpay implements app.PaymentProvider sobre Webpay Plus
// (Transbank), la pasarela de tarjetas más usada en Chile — reemplaza a
// payments/mock cuando hay credenciales configuradas (Etapa 2 §2.3: la
// interfaz ya estaba pensada para esto, ningún caso de uso cambia).
//
// A diferencia del mock, Webpay Plus (Transacción Normal, REST v1.2) es un
// flujo por redirect, no una API pura: el navegador del aportante tiene que
// salir de Juntalo, entrar la tarjeta en la página de Transbank, y volver.
// Por eso el ciclo de vida real tiene 2 pasos server-to-server distintos:
//
//  1. CreateIntent: POST /transactions — arma la transacción y devuelve un
//     token + una URL fija a la que hay que redirigir al navegador (con el
//     token, vía un form POST — Transbank exige esto, un GET no sirve).
//  2. Confirmar (Commit): PUT /transactions/{token} — se llama SOLO cuando
//     Transbank redirige de vuelta al return_url (ver
//     internal/api/handlers/webpay.go), nunca antes.
package webpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

const (
	integrationBaseURL = "https://webpay3gint.transbank.cl"
	productionBaseURL  = "https://webpay3g.transbank.cl"

	transactionsPath = "/rswebpaytransaction/api/webpay/v1.2/transactions"
)

// Config agrupa lo necesario para hablar con Webpay Plus. CommerceCode y
// ApiKey los da Transbank (para "integration" son públicos y compartidos por
// todos los desarrolladores, ver render.yaml); SelfURL arma tanto el
// return_url que Transbank llama al volver como la URL del helper de
// redirect propio (ver Redirect en el handler).
type Config struct {
	CommerceCode string
	APIKey       string
	Environment  string // "integration" | "production"
	SelfURL      string
}

type Provider struct {
	cfg    Config
	client *http.Client
	// testBaseURL solo lo setean los tests (ver webpay_test.go) para apuntar
	// a un httptest.Server en vez de a Transbank real.
	testBaseURL string
}

func New(cfg Config) *Provider {
	return &Provider{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *Provider) baseURL() string {
	if p.testBaseURL != "" {
		return p.testBaseURL
	}
	if p.cfg.Environment == "production" {
		return productionBaseURL
	}
	return integrationBaseURL
}

// buyOrderFrom deriva un buy_order válido (Transbank exige ≤26 caracteres
// alfanuméricos) a partir de la idempotency key del aporte (UUID de 36
// caracteres) — le basta con quitar los guiones y truncar.
func buyOrderFrom(idempotencyKey string) string {
	stripped := strings.ReplaceAll(idempotencyKey, "-", "")
	if len(stripped) > 26 {
		stripped = stripped[:26]
	}
	if stripped == "" {
		stripped = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return stripped
}

type createTransactionRequest struct {
	BuyOrder  string `json:"buy_order"`
	SessionID string `json:"session_id"`
	Amount    int64  `json:"amount"`
	ReturnURL string `json:"return_url"`
}

type createTransactionResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// CreateIntent crea la transacción en Webpay y devuelve una RedirectURL
// propia (no la de Transbank directamente) — Transbank exige un form POST
// con el token, no un GET, así que RedirectURL apunta al helper que arma ese
// form (ver handlers.WebpayHandler.Redirect).
func (p *Provider) CreateIntent(ctx context.Context, req app.IntentRequest) (app.PaymentIntent, error) {
	buyOrder := buyOrderFrom(req.IdempotencyKey)
	body := createTransactionRequest{
		BuyOrder:  buyOrder,
		SessionID: buyOrder,
		Amount:    int64(req.Amount),
		ReturnURL: p.cfg.SelfURL + "/api/v1/webpay/return",
	}

	var resp createTransactionResponse
	if err := p.do(ctx, http.MethodPost, transactionsPath, body, &resp); err != nil {
		return app.PaymentIntent{}, fmt.Errorf("webpay: create transaction: %w", err)
	}

	redirectURL := fmt.Sprintf("%s/api/v1/webpay/redirect?token=%s&init_url=%s",
		p.cfg.SelfURL, resp.Token, url.QueryEscape(resp.URL))

	return app.PaymentIntent{
		ProviderRef: resp.Token,
		Status:      payment.StatusPending,
		RedirectURL: redirectURL,
	}, nil
}

type transactionStatusResponse struct {
	Status       string `json:"status"`
	ResponseCode *int   `json:"response_code"`
}

// GetIntent consulta el estado sin confirmar nada (a diferencia de Commit) —
// útil para depurar, no lo usa el flujo principal (que confirma vía Commit
// en el return handler).
func (p *Provider) GetIntent(ctx context.Context, providerRef string) (app.PaymentIntent, error) {
	var resp transactionStatusResponse
	if err := p.do(ctx, http.MethodGet, transactionsPath+"/"+providerRef, nil, &resp); err != nil {
		return app.PaymentIntent{}, fmt.Errorf("webpay: get transaction: %w", err)
	}
	return app.PaymentIntent{ProviderRef: providerRef, Status: mapStatus(resp.Status, resp.ResponseCode)}, nil
}

// Commit confirma (captura) la transacción — se llama exactamente una vez,
// cuando el navegador vuelve del return_url con el token. No es parte de
// app.PaymentProvider porque ningún otro proveedor tiene este paso; el
// handler de retorno llama este método directo sobre *Provider.
func (p *Provider) Commit(ctx context.Context, token string) (payment.Status, error) {
	var resp transactionStatusResponse
	if err := p.do(ctx, http.MethodPut, transactionsPath+"/"+token, nil, &resp); err != nil {
		return payment.StatusFailed, fmt.Errorf("webpay: commit transaction: %w", err)
	}
	return mapStatus(resp.Status, resp.ResponseCode), nil
}

func mapStatus(status string, responseCode *int) payment.Status {
	if status == "AUTHORIZED" && responseCode != nil && *responseCode == 0 {
		return payment.StatusConfirmed
	}
	if status == "INITIALIZED" {
		return payment.StatusPending
	}
	return payment.StatusFailed
}

func (p *Provider) do(ctx context.Context, method, path string, reqBody, respBody any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		encoded, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, p.baseURL()+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Tbk-Api-Key-Id", p.cfg.CommerceCode)
	httpReq.Header.Set("Tbk-Api-Key-Secret", p.cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webpay responded %d: %s", resp.StatusCode, string(data))
	}
	if respBody != nil && len(data) > 0 {
		if err := json.Unmarshal(data, respBody); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
