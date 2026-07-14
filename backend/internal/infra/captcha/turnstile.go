// Package captcha implements app.CaptchaVerifier sobre Cloudflare Turnstile
// — un token de un solo uso que el frontend obtiene resolviendo el widget,
// y que el backend valida server-side contra la API de Cloudflare antes de
// crear la cuenta (registro es el único endpoint público sin rate limit
// suficiente para frenar creación masiva de cuentas por sí solo).
package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const siteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type TurnstileVerifier struct {
	secretKey string
	client    *http.Client
}

func NewTurnstileVerifier(secretKey string) *TurnstileVerifier {
	return &TurnstileVerifier{secretKey: secretKey, client: &http.Client{Timeout: 5 * time.Second}}
}

type siteverifyResponse struct {
	Success bool `json:"success"`
}

func (v *TurnstileVerifier) Verify(ctx context.Context, token, remoteIP string) (bool, error) {
	if token == "" {
		return false, nil
	}

	form := url.Values{
		"secret":   {v.secretKey},
		"response": {token},
	}
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, siteverifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, fmt.Errorf("turnstile: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("turnstile: request failed: %w", err)
	}
	defer resp.Body.Close()

	var out siteverifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, fmt.Errorf("turnstile: decode response: %w", err)
	}
	return out.Success, nil
}
