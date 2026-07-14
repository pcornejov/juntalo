package webpay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

// newTestProvider points a Provider at a fake Transbank server instead of
// the real integration environment — deterministic, no network dependency,
// covers response shapes que un flujo real completo (browser + tarjeta de
// prueba) no ejercita fácil en CI, en particular el mapeo de AUTHORIZED+0 a
// StatusConfirmed que ya se verificó manualmente contra el sandbox real de
// Transbank (ver sesión de verificación) pero no queda cubierto por un test.
func newTestProvider(t *testing.T, handler http.HandlerFunc) *Provider {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	p := New(Config{CommerceCode: "597055555532", APIKey: "test-key", Environment: "integration", SelfURL: "http://self.example"})
	p.client = srv.Client()
	p.testBaseURL = srv.URL
	return p
}

func TestProvider_CreateIntent_Success(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != transactionsPath {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Tbk-Api-Key-Id") != "597055555532" {
			t.Errorf("missing commerce code header")
		}
		_ = json.NewEncoder(w).Encode(createTransactionResponse{Token: "tok_abc", URL: "https://transbank.example/init"})
	})

	intent, err := p.CreateIntent(context.Background(), app.IntentRequest{IdempotencyKey: "11111111-2222-3333-4444-555555555555", Amount: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Status != payment.StatusPending {
		t.Errorf("got status %s, want pending", intent.Status)
	}
	if intent.ProviderRef != "tok_abc" {
		t.Errorf("got provider ref %q", intent.ProviderRef)
	}
	if intent.RedirectURL == "" {
		t.Error("expected a non-empty RedirectURL")
	}
}

func TestProvider_Commit_Authorized(t *testing.T) {
	responseCode := 0
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(transactionStatusResponse{Status: "AUTHORIZED", ResponseCode: &responseCode})
	})

	status, err := p.Commit(context.Background(), "tok_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != payment.StatusConfirmed {
		t.Errorf("got status %s, want confirmed", status)
	}
}

func TestProvider_Commit_Rejected(t *testing.T) {
	responseCode := -1
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(transactionStatusResponse{Status: "FAILED", ResponseCode: &responseCode})
	})

	status, err := p.Commit(context.Background(), "tok_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != payment.StatusFailed {
		t.Errorf("got status %s, want failed", status)
	}
}

func TestProvider_Commit_HTTPError(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error_message":"boom"}`))
	})

	status, err := p.Commit(context.Background(), "tok_abc")
	if err == nil {
		t.Fatal("expected an error")
	}
	if status != payment.StatusFailed {
		t.Errorf("got status %s, want failed on error", status)
	}
}

func TestProvider_Refund_Success(t *testing.T) {
	responseCode := 0
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(refundResponse{Type: "REVERSED", ResponseCode: &responseCode, AuthorizationCode: "auth123"})
	})

	result, err := p.Refund(context.Background(), "tok_abc", money.CLP(500))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Amount != money.CLP(500) {
		t.Errorf("got amount %d, want 500", result.Amount)
	}
	if result.ProviderRef != "auth123" {
		t.Errorf("got provider ref %q", result.ProviderRef)
	}
}

func TestProvider_Refund_Rejected(t *testing.T) {
	responseCode := -1
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(refundResponse{Type: "NULLIFIED", ResponseCode: &responseCode})
	})

	if _, err := p.Refund(context.Background(), "tok_abc", money.CLP(500)); err == nil {
		t.Fatal("expected an error for a rejected refund")
	}
}
