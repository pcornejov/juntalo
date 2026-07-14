package auth

import (
	"context"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

var errCaptchaFailed = apperr.New("captcha_failed", "No pudimos verificar que no eres un robot, intenta de nuevo")

type RegisterService struct {
	repo    app.AuthRepository
	hasher  app.PasswordHasher
	captcha app.CaptchaVerifier
}

func NewRegisterService(repo app.AuthRepository, hasher app.PasswordHasher, captcha app.CaptchaVerifier) *RegisterService {
	return &RegisterService{repo: repo, hasher: hasher, captcha: captcha}
}

func (s *RegisterService) Register(ctx context.Context, email, fullName, password, captchaToken, remoteIP string) (identity.User, identity.Organization, error) {
	if err := identity.ValidatePassword(password); err != nil {
		return identity.User{}, identity.Organization{}, err
	}

	ok, err := s.captcha.Verify(ctx, captchaToken, remoteIP)
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}
	if !ok {
		return identity.User{}, identity.Organization{}, errCaptchaFailed
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}

	return s.repo.Register(ctx, app.RegisterInput{
		Email:        email,
		FullName:     fullName,
		PasswordHash: hash,
	})
}
