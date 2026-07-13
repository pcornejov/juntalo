package mock

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type webhookPayload struct {
	ProviderRef string `json:"provider_ref"`
	Event       string `json:"event"`
}

// Sign computes the HMAC-SHA256 signature the webhook handler verifies —
// exported so the handler and this package share one implementation
// (Etapa 4 §5: el pipeline de verificación queda ensayado).
func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func sendWebhook(url, secret, providerRef, event string) {
	body, err := json.Marshal(webhookPayload{ProviderRef: providerRef, Event: event})
	if err != nil {
		log.Printf("mock provider: marshal webhook payload: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("mock provider: build webhook request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", Sign(secret, body))

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("mock provider: webhook call failed: %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		log.Printf("mock provider: webhook responded with status %d", resp.StatusCode)
	}
}
