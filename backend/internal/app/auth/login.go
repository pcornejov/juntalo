package auth

import (
	"context"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

var errInvalidCredentials = apperr.New("invalid_credentials", "Email o contraseña incorrectos")

type LoginService struct {
	users  app.UserRepository
	auth   app.AuthRepository
	orgs   app.OrganizationRepository
	hasher app.PasswordHasher
	// dummyHash: un hash argon2id válido sobre el que corremos Verify cuando
	// el email no existe, para que el tiempo de respuesta no delate si una
	// cuenta existe o no (auditoría de seguridad — antes se retornaba de
	// inmediato sin correr el hash, dejando un canal de timing medible en un
	// endpoint sin rate limit).
	dummyHash string
}

func NewLoginService(users app.UserRepository, auth app.AuthRepository, orgs app.OrganizationRepository, hasher app.PasswordHasher) *LoginService {
	dummyHash, _ := hasher.Hash("dummy-password-for-timing-safety")
	return &LoginService{users: users, auth: auth, orgs: orgs, hasher: hasher, dummyHash: dummyHash}
}

// Login never reveals whether the email exists (Etapa 4 §2): every failure
// returns the same generic invalid_credentials error.
func (s *LoginService) Login(ctx context.Context, email, password string) (identity.User, identity.Organization, error) {
	user, found, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}
	if !found {
		s.hasher.Verify(password, s.dummyHash)
		return identity.User{}, identity.Organization{}, errInvalidCredentials
	}

	hash, err := s.auth.GetPasswordHash(ctx, user.ID)
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}
	if !s.hasher.Verify(password, hash) {
		return identity.User{}, identity.Organization{}, errInvalidCredentials
	}

	org, err := s.orgs.GetPersonalByUserID(ctx, user.ID)
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}

	return user, org, nil
}
