package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
)

const EmailVerificationTokenTTL = 24 * time.Hour

var ErrInvalidVerificationToken = apperr.New("invalid_verification_token", "El link de verificación no es válido o ya expiró")

type EmailVerificationService struct {
	users  app.UserRepository
	tokens app.EmailVerificationRepository
}

func NewEmailVerificationService(users app.UserRepository, tokens app.EmailVerificationRepository) *EmailVerificationService {
	return &EmailVerificationService{users: users, tokens: tokens}
}

// IssueToken genera y persiste un token de verificación para userID — se
// llama justo después de registrarse, y también desde "reenviar
// verificación" para quien no llegó a hacer click a tiempo.
func (s *EmailVerificationService) IssueToken(ctx context.Context, userID uuid.UUID) (string, error) {
	raw, err := randomToken()
	if err != nil {
		return "", err
	}
	if err := s.tokens.Create(ctx, userID, hashToken(raw), time.Now().Add(EmailVerificationTokenTTL)); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *EmailVerificationService) Verify(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	userID, found, err := s.tokens.GetUserIDByValidHash(ctx, hash)
	if err != nil {
		return err
	}
	if !found {
		return ErrInvalidVerificationToken
	}
	if err := s.users.MarkEmailVerified(ctx, userID); err != nil {
		return err
	}
	return s.tokens.MarkUsed(ctx, hash)
}
