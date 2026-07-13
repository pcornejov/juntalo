package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTSigner_SignAndParse(t *testing.T) {
	signer := NewJWTSigner("test-secret")
	userID := uuid.New()

	token, err := signer.Sign(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	got, err := signer.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got != userID {
		t.Errorf("got %v, want %v", got, userID)
	}
}

func TestJWTSigner_RejectsExpired(t *testing.T) {
	signer := NewJWTSigner("test-secret")
	token, err := signer.Sign(uuid.New(), -1*time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := signer.Parse(token); err == nil {
		t.Error("expected expired token to fail parsing")
	}
}

func TestJWTSigner_RejectsWrongSecret(t *testing.T) {
	token, err := NewJWTSigner("secret-a").Sign(uuid.New(), 15*time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := NewJWTSigner("secret-b").Parse(token); err == nil {
		t.Error("expected token signed with a different secret to fail parsing")
	}
}
