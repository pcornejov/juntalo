package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

type fakeGoogleVerifier struct {
	claims map[string]app.GoogleClaims // idToken -> claims
}

func (f *fakeGoogleVerifier) Verify(_ context.Context, idToken string) (app.GoogleClaims, error) {
	c, ok := f.claims[idToken]
	if !ok {
		return app.GoogleClaims{}, errors.New("invalid token")
	}
	return c, nil
}

func TestGoogleLoginService_NewAccount(t *testing.T) {
	authRepo := newFakeAuthRepo()
	users := &fakeUserRepo{byEmail: map[string]identity.User{}}
	orgs := &fakeOrgRepo{byUser: map[uuid.UUID]identity.Organization{}}
	verifier := &fakeGoogleVerifier{claims: map[string]app.GoogleClaims{
		"tok": {Subject: "sub-1", Email: "nueva@example.com", EmailVerified: true, FullName: "Nueva Vecina"},
	}}
	svc := NewGoogleLoginService(authRepo, users, orgs, verifier)

	user, org, err := svc.Login(context.Background(), "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "nueva@example.com" || !user.EmailVerified {
		t.Errorf("expected a new verified user, got %+v", user)
	}
	if org.Kind != "personal" {
		t.Errorf("expected a personal organization, got %+v", org)
	}
	if _, found, _ := authRepo.GetUserIDByGoogleSubject(context.Background(), "sub-1"); !found {
		t.Error("expected the google subject to be linked to the new user")
	}
}

func TestGoogleLoginService_ExistingGoogleIdentity(t *testing.T) {
	authRepo := newFakeAuthRepo()
	userID := uuid.New()
	authRepo.googleSubjects["sub-1"] = userID

	existingUser := identity.User{ID: userID, Email: "ya@example.com", FullName: "Ya Existe", Status: "active"}
	users := &fakeUserRepo{byEmail: map[string]identity.User{existingUser.Email: existingUser}}
	orgs := &fakeOrgRepo{byUser: map[uuid.UUID]identity.Organization{userID: {ID: uuid.New(), Kind: "personal"}}}
	verifier := &fakeGoogleVerifier{claims: map[string]app.GoogleClaims{
		"tok": {Subject: "sub-1", Email: "ya@example.com", EmailVerified: true, FullName: "Ya Existe"},
	}}
	svc := NewGoogleLoginService(authRepo, users, orgs, verifier)

	user, _, err := svc.Login(context.Background(), "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != userID {
		t.Errorf("expected the already-linked user, got %v", user.ID)
	}
}

func TestGoogleLoginService_LinksToExistingPasswordAccount(t *testing.T) {
	authRepo := newFakeAuthRepo()
	userID := uuid.New()
	existingUser := identity.User{ID: userID, Email: "conpassword@example.com", FullName: "Con Password", Status: "active"}
	users := &fakeUserRepo{byEmail: map[string]identity.User{existingUser.Email: existingUser}}
	orgs := &fakeOrgRepo{byUser: map[uuid.UUID]identity.Organization{userID: {ID: uuid.New(), Kind: "personal"}}}
	verifier := &fakeGoogleVerifier{claims: map[string]app.GoogleClaims{
		"tok": {Subject: "sub-new", Email: "conpassword@example.com", EmailVerified: true, FullName: "Con Password"},
	}}
	svc := NewGoogleLoginService(authRepo, users, orgs, verifier)

	user, _, err := svc.Login(context.Background(), "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != userID {
		t.Errorf("expected to link into the existing account, got a different user %v", user.ID)
	}
	if linked, found, _ := authRepo.GetUserIDByGoogleSubject(context.Background(), "sub-new"); !found || linked != userID {
		t.Error("expected the google subject to now be linked to the existing user")
	}
}

func TestGoogleLoginService_RejectsUnverifiedEmail(t *testing.T) {
	authRepo := newFakeAuthRepo()
	users := &fakeUserRepo{byEmail: map[string]identity.User{}}
	orgs := &fakeOrgRepo{byUser: map[uuid.UUID]identity.Organization{}}
	verifier := &fakeGoogleVerifier{claims: map[string]app.GoogleClaims{
		"tok": {Subject: "sub-1", Email: "sinverificar@example.com", EmailVerified: false, FullName: "Sin Verificar"},
	}}
	svc := NewGoogleLoginService(authRepo, users, orgs, verifier)

	_, _, err := svc.Login(context.Background(), "tok")
	if !apperr.Is(err, "google_email_not_verified") {
		t.Fatalf("expected google_email_not_verified, got %v", err)
	}
}

func TestGoogleLoginService_RejectsInvalidToken(t *testing.T) {
	authRepo := newFakeAuthRepo()
	users := &fakeUserRepo{byEmail: map[string]identity.User{}}
	orgs := &fakeOrgRepo{byUser: map[uuid.UUID]identity.Organization{}}
	verifier := &fakeGoogleVerifier{claims: map[string]app.GoogleClaims{}}
	svc := NewGoogleLoginService(authRepo, users, orgs, verifier)

	_, _, err := svc.Login(context.Background(), "not-a-real-token")
	if !apperr.Is(err, "invalid_google_token") {
		t.Fatalf("expected invalid_google_token, got %v", err)
	}
}
