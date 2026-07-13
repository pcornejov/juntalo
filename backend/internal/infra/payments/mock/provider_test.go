package mock

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

func TestProvider_InstantMode(t *testing.T) {
	p := NewProvider(ModeInstant, "http://unused", "secret")
	intent, err := p.CreateIntent(context.Background(), app.IntentRequest{IdempotencyKey: "key-1", Amount: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Status != payment.StatusConfirmed {
		t.Errorf("got status %s, want confirmed", intent.Status)
	}
}

func TestProvider_FailMode(t *testing.T) {
	p := NewProvider(ModeFail, "http://unused", "secret")
	intent, err := p.CreateIntent(context.Background(), app.IntentRequest{IdempotencyKey: "key-1", Amount: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Status != payment.StatusFailed {
		t.Errorf("got status %s, want failed", intent.Status)
	}
}

// TestProvider_SameIdempotencyKeyReturnsSameIntent exercises the requirement
// from Etapa 1 §2: reintentos/duplicados con la misma key no crean un intent
// nuevo — la idempotencia debe llegar probada al adaptador, no solo al caso de uso.
func TestProvider_SameIdempotencyKeyReturnsSameIntent(t *testing.T) {
	p := NewProvider(ModeInstant, "http://unused", "secret")
	first, err := p.CreateIntent(context.Background(), app.IntentRequest{IdempotencyKey: "same-key", Amount: 1000})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	second, err := p.CreateIntent(context.Background(), app.IntentRequest{IdempotencyKey: "same-key", Amount: 1000})
	if err != nil {
		t.Fatalf("second call (retry): %v", err)
	}
	if first.ProviderRef != second.ProviderRef {
		t.Errorf("expected same ProviderRef for a retried idempotency key, got %q vs %q", first.ProviderRef, second.ProviderRef)
	}
}

// TestProvider_DeferredModeCallsWebhook verifies the self-invoked, signed
// webhook call that ensaya el pipeline real de confirmación asíncrona
// (Etapa 2 §2.3, Etapa 4 §5).
func TestProvider_DeferredModeCallsWebhook(t *testing.T) {
	received := make(chan struct {
		event     string
		signature string
		body      []byte
	}, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received <- struct {
			event     string
			signature string
			body      []byte
		}{signature: r.Header.Get("X-Signature"), body: body}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	secret := "test-secret"
	p := NewProvider(ModeDeferred, server.URL, secret)

	intent, err := p.CreateIntent(context.Background(), app.IntentRequest{IdempotencyKey: "deferred-key", Amount: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Status != payment.StatusPending {
		t.Errorf("deferred mode should return pending initially, got %s", intent.Status)
	}

	select {
	case msg := <-received:
		expectedSig := Sign(secret, msg.body)
		if msg.signature != expectedSig {
			t.Errorf("signature mismatch: got %s, want %s", msg.signature, expectedSig)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for the mock provider to self-invoke the webhook")
	}
}
