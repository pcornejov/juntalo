// Package payment models the financial core of Juntalo: payments, their state
// machine, and commission calculation (Etapa 3 §5 — el corazón del producto).
package payment

import (
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

type Payment struct {
	ID                    uuid.UUID
	ContributionID        uuid.UUID
	IdempotencyKey        string
	Provider              string
	ProviderRef           string
	Status                Status
	AmountGross           money.CLP
	CommissionRateApplied float64
	CommissionAmount      money.CLP
	AmountNet             money.CLP
	PayeeSnapshot         map[string]string
	ConfirmedAt           *time.Time
	FailedAt              *time.Time
	CreatedAt             time.Time
}
