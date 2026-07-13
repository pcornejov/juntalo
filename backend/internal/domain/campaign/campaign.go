// Package campaign models campaigns: entity, status transitions, and the
// declarative type registry (Etapa 3 §3, Etapa 1 riesgo 1).
package campaign

import (
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

type Campaign struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	TypeKey        TypeKey
	Title          string
	Slug           string
	Description    string
	CoverFileID    *uuid.UUID
	GoalAmount     *money.CLP
	Status         Status
	StartsAt       *time.Time
	EndsAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Totals is the read model backed by the campaign_totals view (Etapa 3 §5:
// agregación, no contador — sin condiciones de carrera).
type Totals struct {
	RaisedGross      money.CLP
	RaisedNetApprox  money.CLP
	ContributorCount int64
}
