// Package email implements app.EmailSender: Resend como proveedor real, y un
// no-op para ambientes sin RESEND_API_KEY configurada (Etapa 2 §2.1 — mismo
// patrón que PaymentProvider: la interfaz no cambia, la implementación sí).
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const resendAPIURL = "https://api.resend.com/emails"

type ResendSender struct {
	apiKey     string
	from       string
	httpClient *http.Client
}

func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{
		apiKey:     apiKey,
		from:       from,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *ResendSender) Send(ctx context.Context, to, subject, htmlBody string) error {
	body, err := json.Marshal(resendRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
	})
	if err != nil {
		return fmt.Errorf("email: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("email: send: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= 300 {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("email: resend respondió %d: %s", res.StatusCode, string(respBody))
	}
	return nil
}
