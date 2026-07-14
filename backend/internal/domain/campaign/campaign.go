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
	Category       Category
	Title          string
	Slug           string
	Description    string
	CoverFileID    *uuid.UUID
	// VideoURL es un link externo a YouTube o Vimeo — Juntalo no aloja video
	// propio (evita el costo de storage y de servir range requests para seek).
	VideoURL   *string
	GoalAmount *money.CLP
	Status     Status
	StartsAt   *time.Time
	EndsAt     *time.Time
	// PublishAt: si está seteada y la campaña sigue en draft, el scheduler en
	// background la publica automáticamente al llegar esa fecha (Etapa 4).
	PublishAt *time.Time
	// Campos exclusivos de TypeRaffle: precio fijo por número, rango total, y
	// el número ganador que el organizador registra tras el sorteo externo
	// (Juntalo no sortea nada dentro de la plataforma).
	RaffleUnitPrice     *money.CLP
	RaffleTotalNumbers  *int
	RaffleWinningNumber *int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Totals is the read model backed by the campaign_totals view (Etapa 3 §5:
// agregación, no contador — sin condiciones de carrera).
type Totals struct {
	RaisedGross      money.CLP
	RaisedNetApprox  money.CLP
	ContributorCount int64
}
