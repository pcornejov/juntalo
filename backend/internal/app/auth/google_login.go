package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

var (
	errInvalidGoogleToken     = apperr.New("invalid_google_token", "No se pudo verificar la sesión de Google")
	errGoogleEmailNotVerified = apperr.New("google_email_not_verified", "Tu cuenta de Google no tiene el email verificado")
)

type GoogleLoginService struct {
	auth     app.AuthRepository
	users    app.UserRepository
	orgs     app.OrganizationRepository
	verifier app.GoogleTokenVerifier
}

func NewGoogleLoginService(auth app.AuthRepository, users app.UserRepository, orgs app.OrganizationRepository, verifier app.GoogleTokenVerifier) *GoogleLoginService {
	return &GoogleLoginService{auth: auth, users: users, orgs: orgs, verifier: verifier}
}

// Login resuelve un ID token de Google en tres pasos posibles (Etapa 1: "se
// modela pensando en esto, pero no se implementa" — esto es esa implementación):
//  1. Ya existe una identidad de Google con este sub → login directo.
//  2. No existe identidad de Google, pero el email ya tiene cuenta (creada
//     con password) → se vincula la identidad de Google a esa cuenta. Solo se
//     hace si Google confirma email_verified=true — es la garantía que evita
//     que alguien tome una cuenta ajena solo por compartir el email.
//  3. Ni identidad ni cuenta existentes → se crea una cuenta nueva, igual que
//     Register pero sin password.
func (s *GoogleLoginService) Login(ctx context.Context, idToken string) (identity.User, identity.Organization, error) {
	claims, err := s.verifier.Verify(ctx, idToken)
	if err != nil {
		return identity.User{}, identity.Organization{}, errInvalidGoogleToken
	}
	if !claims.EmailVerified {
		return identity.User{}, identity.Organization{}, errGoogleEmailNotVerified
	}

	if userID, found, err := s.auth.GetUserIDByGoogleSubject(ctx, claims.Subject); err != nil {
		return identity.User{}, identity.Organization{}, err
	} else if found {
		return s.sessionFor(ctx, userID)
	}

	if existing, found, err := s.users.GetByEmail(ctx, claims.Email); err != nil {
		return identity.User{}, identity.Organization{}, err
	} else if found {
		if err := s.auth.LinkGoogleIdentity(ctx, existing.ID, claims.Subject); err != nil {
			return identity.User{}, identity.Organization{}, err
		}
		org, err := s.orgs.GetPersonalByUserID(ctx, existing.ID)
		if err != nil {
			return identity.User{}, identity.Organization{}, err
		}
		return existing, org, nil
	}

	user, org, err := s.auth.RegisterGoogle(ctx, app.RegisterGoogleInput{
		Email:    claims.Email,
		FullName: claims.FullName,
		Subject:  claims.Subject,
	})
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}
	// Google ya verificó el email — no tiene sentido pedirle a este usuario
	// que lo vuelva a verificar por el flujo de link-por-email de Juntalo.
	// Best-effort, igual que el envío de email de verificación en Register:
	// si falla, la cuenta queda creada y verificable a mano después.
	_ = s.users.MarkEmailVerified(ctx, user.ID)
	user.EmailVerified = true
	return user, org, nil
}

func (s *GoogleLoginService) sessionFor(ctx context.Context, userID uuid.UUID) (identity.User, identity.Organization, error) {
	user, found, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}
	if !found {
		return identity.User{}, identity.Organization{}, errInvalidGoogleToken
	}
	org, err := s.orgs.GetPersonalByUserID(ctx, userID)
	if err != nil {
		return identity.User{}, identity.Organization{}, err
	}
	return user, org, nil
}
