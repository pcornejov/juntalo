package dto

import (
	"github.com/go-playground/validator/v10"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
)

var validate = validator.New()

// Validate applies struct tags at the API edge (Etapa 2 §2.5). Business invariants
// are re-validated in domain — this is only the first line of defense.
func Validate(v any) error {
	if err := validate.Struct(v); err != nil {
		return apperr.New("validation_failed", "Los datos enviados no son válidos")
	}
	return nil
}
