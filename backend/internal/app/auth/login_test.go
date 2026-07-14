package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

type fakeUserRepo struct {
	byEmail map[string]identity.User
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (identity.User, bool, error) {
	u, ok := f.byEmail[email]
	return u, ok, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (identity.User, bool, error) {
	for _, u := range f.byEmail {
		if u.ID == id {
			return u, true, nil
		}
	}
	return identity.User{}, false, nil
}

func (f *fakeUserRepo) MarkEmailVerified(_ context.Context, _ uuid.UUID) error {
	return nil
}

type fakeOrgRepo struct {
	byUser map[uuid.UUID]identity.Organization
}

func (f *fakeOrgRepo) GetPersonalByUserID(_ context.Context, userID uuid.UUID) (identity.Organization, error) {
	return f.byUser[userID], nil
}

func (f *fakeOrgRepo) GetByID(_ context.Context, id uuid.UUID) (identity.Organization, error) {
	for _, o := range f.byUser {
		if o.ID == id {
			return o, nil
		}
	}
	return identity.Organization{}, nil
}

func (f *fakeOrgRepo) GetOwnerEmail(context.Context, uuid.UUID) (string, string, error) {
	return "", "", nil
}

func (f *fakeOrgRepo) GetOwnerInfo(context.Context, uuid.UUID) (string, bool, error) {
	return "", false, nil
}

func setupLoginFixture(t *testing.T) (*LoginService, identity.User) {
	t.Helper()
	authRepo := newFakeAuthRepo()
	hasher := fakeHasher{}

	userID := uuid.New()
	user := identity.User{ID: userID, Email: "ana@example.com", FullName: "Ana", Status: "active"}
	authRepo.emails[user.Email] = true
	authRepo.passwords[userID] = "hashed:password123"

	users := &fakeUserRepo{byEmail: map[string]identity.User{user.Email: user}}
	orgs := &fakeOrgRepo{byUser: map[uuid.UUID]identity.Organization{userID: {ID: uuid.New(), Kind: "personal"}}}

	return NewLoginService(users, authRepo, orgs, hasher), user
}

func TestLoginService_Success(t *testing.T) {
	svc, user := setupLoginFixture(t)

	got, org, err := svc.Login(context.Background(), user.Email, "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("got user %v, want %v", got.ID, user.ID)
	}
	if org.Kind != "personal" {
		t.Errorf("got org kind %q", org.Kind)
	}
}

func TestLoginService_WrongPassword(t *testing.T) {
	svc, user := setupLoginFixture(t)

	_, _, err := svc.Login(context.Background(), user.Email, "wrong-password")
	if !apperr.Is(err, "invalid_credentials") {
		t.Fatalf("expected invalid_credentials, got %v", err)
	}
}

func TestLoginService_UnknownEmail(t *testing.T) {
	svc, _ := setupLoginFixture(t)

	_, _, err := svc.Login(context.Background(), "nadie@example.com", "password123")
	if !apperr.Is(err, "invalid_credentials") {
		t.Fatalf("expected invalid_credentials (no filtrar existencia), got %v", err)
	}
}
