package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

const RefreshTokenTTL = 30 * 24 * time.Hour

var ErrSessionExpired = apperr.New("session_expired", "La sesión expiró, inicia sesión nuevamente")

type RefreshService struct {
	tokens app.RefreshTokenRepository
}

func NewRefreshService(tokens app.RefreshTokenRepository) *RefreshService {
	return &RefreshService{tokens: tokens}
}

// IssueRefreshToken creates a new refresh token for userID and returns the raw
// (unhashed) value that goes in the httpOnly cookie — only the hash is persisted.
func (s *RefreshService) IssueRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	raw, err := randomToken()
	if err != nil {
		return "", err
	}
	err = s.tokens.Create(ctx, identity.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
	})
	return raw, err
}

// Rotate validates the raw refresh token, revokes it, and issues a new one
// (rotación, Etapa 3 §2 — un token usado no puede reutilizarse).
func (s *RefreshService) Rotate(ctx context.Context, rawToken string) (uuid.UUID, string, error) {
	stored, found, err := s.tokens.GetValidByHash(ctx, hashToken(rawToken))
	if err != nil {
		return uuid.Nil, "", err
	}
	if !found || !stored.IsValid(time.Now()) {
		return uuid.Nil, "", ErrSessionExpired
	}

	if err := s.tokens.Revoke(ctx, stored.ID); err != nil {
		return uuid.Nil, "", err
	}

	newRaw, err := s.IssueRefreshToken(ctx, stored.UserID)
	if err != nil {
		return uuid.Nil, "", err
	}
	return stored.UserID, newRaw, nil
}

func (s *RefreshService) Revoke(ctx context.Context, rawToken string) error {
	stored, found, err := s.tokens.GetValidByHash(ctx, hashToken(rawToken))
	if err != nil {
		return err
	}
	if !found {
		return nil // logout de un token ya inválido no es un error
	}
	return s.tokens.Revoke(ctx, stored.ID)
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
