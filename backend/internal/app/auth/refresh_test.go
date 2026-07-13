package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

type fakeRefreshRepo struct {
	byHash map[string]identity.RefreshToken
}

func newFakeRefreshRepo() *fakeRefreshRepo {
	return &fakeRefreshRepo{byHash: map[string]identity.RefreshToken{}}
}

func (f *fakeRefreshRepo) Create(_ context.Context, t identity.RefreshToken) error {
	f.byHash[t.TokenHash] = t
	return nil
}

func (f *fakeRefreshRepo) GetValidByHash(_ context.Context, tokenHash string) (identity.RefreshToken, bool, error) {
	t, ok := f.byHash[tokenHash]
	return t, ok, nil
}

func (f *fakeRefreshRepo) Revoke(_ context.Context, id uuid.UUID) error {
	for hash, t := range f.byHash {
		if t.ID == id {
			now := time.Now()
			t.RevokedAt = &now
			f.byHash[hash] = t
		}
	}
	return nil
}

func TestRefreshService_RotateInvalidatesOldToken(t *testing.T) {
	repo := newFakeRefreshRepo()
	svc := NewRefreshService(repo)
	userID := uuid.New()

	raw, err := svc.IssueRefreshToken(context.Background(), userID)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	gotUserID, newRaw, err := svc.Rotate(context.Background(), raw)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("got userID %v, want %v", gotUserID, userID)
	}
	if newRaw == raw {
		t.Error("rotate must issue a different token")
	}

	// El token original ya no es válido: reintentar rotarlo debe fallar.
	_, _, err = svc.Rotate(context.Background(), raw)
	if !apperr.Is(err, "session_expired") {
		t.Fatalf("expected session_expired reusing a rotated token, got %v", err)
	}
}

func TestRefreshService_RevokeUnknownTokenIsNoop(t *testing.T) {
	repo := newFakeRefreshRepo()
	svc := NewRefreshService(repo)

	if err := svc.Revoke(context.Background(), "never-issued"); err != nil {
		t.Fatalf("logout of unknown token should not error, got %v", err)
	}
}
