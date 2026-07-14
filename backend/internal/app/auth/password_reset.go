package auth

import (
	"context"
	"time"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

const PasswordResetTokenTTL = 1 * time.Hour

var ErrInvalidResetToken = apperr.New("invalid_reset_token", "El link para restablecer tu contraseña no es válido o ya expiró")

type ForgotPasswordService struct {
	users  app.UserRepository
	resets app.PasswordResetRepository
}

func NewForgotPasswordService(users app.UserRepository, resets app.PasswordResetRepository) *ForgotPasswordService {
	return &ForgotPasswordService{users: users, resets: resets}
}

// RequestReset devuelve el token crudo solo si el email existe; si no
// existe, devuelve "" sin error — el handler responde el mismo mensaje
// genérico en ambos casos para no filtrar qué emails están registrados
// (Etapa 1 riesgo de enumeración de usuarios).
func (s *ForgotPasswordService) RequestReset(ctx context.Context, email string) (string, error) {
	user, found, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if !found {
		return "", nil
	}

	raw, err := randomToken()
	if err != nil {
		return "", err
	}
	if err := s.resets.Create(ctx, user.ID, hashToken(raw), time.Now().Add(PasswordResetTokenTTL)); err != nil {
		return "", err
	}
	return raw, nil
}

type ResetPasswordService struct {
	resets app.PasswordResetRepository
	auth   app.AuthRepository
	hasher app.PasswordHasher
}

func NewResetPasswordService(resets app.PasswordResetRepository, auth app.AuthRepository, hasher app.PasswordHasher) *ResetPasswordService {
	return &ResetPasswordService{resets: resets, auth: auth, hasher: hasher}
}

func (s *ResetPasswordService) Reset(ctx context.Context, rawToken, newPassword string) error {
	if err := identity.ValidatePassword(newPassword); err != nil {
		return err
	}

	hash := hashToken(rawToken)
	userID, found, err := s.resets.GetUserIDByValidHash(ctx, hash)
	if err != nil {
		return err
	}
	if !found {
		return ErrInvalidResetToken
	}

	newHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	if err := s.auth.UpdatePasswordHash(ctx, userID, newHash); err != nil {
		return err
	}
	return s.resets.MarkUsed(ctx, hash)
}
