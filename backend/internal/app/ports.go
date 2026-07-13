// Package app contains use cases (casos de uso) that orchestrate domain logic and
// infrastructure ports. All repository/adapter interfaces live in this single file
// (Etapa 5 §2: "se ve el contrato de persistencia completo de un vistazo").
package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

// RegisterInput carries everything AuthRepository.Register needs to run the
// user+identity+organization+membership transaction (Etapa 3 §2).
type RegisterInput struct {
	Email        string
	FullName     string
	PasswordHash string
}

type AuthRepository interface {
	// Register creates user + password identity + personal organization + owner
	// membership in a single transaction. Returns apperr "email_already_registered"
	// if the email is taken.
	Register(ctx context.Context, in RegisterInput) (identity.User, identity.Organization, error)
	GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error)
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (identity.User, bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (identity.User, bool, error)
}

type OrganizationRepository interface {
	GetPersonalByUserID(ctx context.Context, userID uuid.UUID) (identity.Organization, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t identity.RefreshToken) error
	GetValidByHash(ctx context.Context, tokenHash string) (identity.RefreshToken, bool, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}

// PasswordHasher isolates the hashing algorithm (argon2id) from the use cases that need it.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// TokenSigner issues and parses short-lived JWT access tokens.
type TokenSigner interface {
	Sign(userID uuid.UUID, ttl time.Duration) (string, error)
	Parse(token string) (uuid.UUID, error)
}
