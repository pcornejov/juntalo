// Package identity models users, organizations and credentials (Etapa 3 §2).
package identity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	FullName  string
	Status    string
	CreatedAt time.Time
}

type Organization struct {
	ID             uuid.UUID
	Name           string
	Kind           string
	CommissionRate float64
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func (t RefreshToken) IsValid(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}
