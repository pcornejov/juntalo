package contribution

import (
	"strings"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

var (
	errAmountOutOfRange = apperr.New("amount_out_of_range", "El monto debe ser mayor a cero")
	errNameRequired     = apperr.New("validation_failed", "El nombre es obligatorio")
)

// maxAmount is a sanity ceiling against fat-finger / abuse in the MVP —
// $50.000.000 CLP por aporte, revisable cuando haya datos reales de uso.
const maxAmount = money.CLP(50_000_000)

func ValidateAmount(amount money.CLP) error {
	if !amount.IsPositive() || amount > maxAmount {
		return errAmountOutOfRange
	}
	return nil
}

func ValidateContributorName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errNameRequired
	}
	return nil
}
