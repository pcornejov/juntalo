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

type fakeAuthRepo struct {
	emails    map[string]bool
	passwords map[uuid.UUID]string
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{emails: map[string]bool{}, passwords: map[uuid.UUID]string{}}
}

func (f *fakeAuthRepo) Register(_ context.Context, in app.RegisterInput) (identity.User, identity.Organization, error) {
	if f.emails[in.Email] {
		return identity.User{}, identity.Organization{}, apperr.New("email_already_registered", "ya existe")
	}
	f.emails[in.Email] = true
	userID := uuid.New()
	f.passwords[userID] = in.PasswordHash
	user := identity.User{ID: userID, Email: in.Email, FullName: in.FullName, Status: "active"}
	org := identity.Organization{ID: uuid.New(), Name: "org de " + in.FullName, Kind: "personal"}
	return user, org, nil
}

func (f *fakeAuthRepo) GetPasswordHash(_ context.Context, userID uuid.UUID) (string, error) {
	hash, ok := f.passwords[userID]
	if !ok {
		return "", errors.New("not found")
	}
	return hash, nil
}

func (f *fakeAuthRepo) UpdatePasswordHash(_ context.Context, userID uuid.UUID, newHash string) error {
	f.passwords[userID] = newHash
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (fakeHasher) Verify(password, hash string) bool    { return hash == "hashed:"+password }

// fakeCaptchaVerifier por defecto aprueba siempre (equivalente al NoopVerifier
// real) — approve=false simula un token de Turnstile rechazado.
type fakeCaptchaVerifier struct {
	approve bool
}

func (f fakeCaptchaVerifier) Verify(context.Context, string, string) (bool, error) {
	return f.approve, nil
}

func TestRegisterService_Register(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewRegisterService(repo, fakeHasher{}, fakeCaptchaVerifier{approve: true})

	user, org, err := svc.Register(context.Background(), "ana@example.com", "Ana Pérez", "password123", "token", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "ana@example.com" {
		t.Errorf("got email %q", user.Email)
	}
	if org.Kind != "personal" {
		t.Errorf("got org kind %q, want personal", org.Kind)
	}
}

func TestRegisterService_WeakPassword(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewRegisterService(repo, fakeHasher{}, fakeCaptchaVerifier{approve: true})

	_, _, err := svc.Register(context.Background(), "ana@example.com", "Ana", "short", "token", "1.2.3.4")
	if !apperr.Is(err, "weak_password") {
		t.Fatalf("expected weak_password error, got %v", err)
	}
}

func TestRegisterService_DuplicateEmail(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewRegisterService(repo, fakeHasher{}, fakeCaptchaVerifier{approve: true})

	_, _, err := svc.Register(context.Background(), "ana@example.com", "Ana", "password123", "token", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error on first register: %v", err)
	}

	_, _, err = svc.Register(context.Background(), "ana@example.com", "Otra Ana", "password456", "token", "1.2.3.4")
	if !apperr.Is(err, "email_already_registered") {
		t.Fatalf("expected email_already_registered, got %v", err)
	}
}

func TestRegisterService_CaptchaRejected(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewRegisterService(repo, fakeHasher{}, fakeCaptchaVerifier{approve: false})

	_, _, err := svc.Register(context.Background(), "ana@example.com", "Ana", "password123", "bad-token", "1.2.3.4")
	if !apperr.Is(err, "captcha_failed") {
		t.Fatalf("expected captcha_failed, got %v", err)
	}
}
