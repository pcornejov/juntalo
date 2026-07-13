package auth

import (
	"context"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

type RegisterService struct {
	repo   app.AuthRepository
	hasher app.PasswordHasher
}

func NewRegisterService(repo app.AuthRepository, hasher app.PasswordHasher) *RegisterService {
	return &RegisterService{repo: repo, hasher: hasher}
}

func (s *RegisterService) Register(ctx context.Context, email, fullName, password string) (identity.User, identity.Organization, error) {
	if err := identity.ValidatePassword(password); err != nil {
		return identity.User{}, identity.Organization{}, err
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
