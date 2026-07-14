// Package contribution models the business fact of contributing to a campaign,
// deliberately separate from payment (Etapa 3 §4): la contribución es el hecho
// de negocio, el pago es la transacción financiera que la respalda.
package contribution

import (
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusFailed    Status = "failed"
	StatusRefunded  Status = "refunded"
)

type Contributor struct {
	ID       uuid.UUID
	FullName string
	Email    string
	Phone    string
}

type Contribution struct {
	ID            uuid.UUID
	CampaignID    uuid.UUID
	ContributorID uuid.UUID
	Amount        money.CLP
	IsAnonymous   bool
	Message       string
	Status        Status
	// RaffleNumber: número asignado a esta contribución cuando la campaña es
	// de tipo "raffle" — nil para el resto de los tipos.
	RaffleNumber *int
	CreatedAt    time.Time
}
