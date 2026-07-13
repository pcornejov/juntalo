package identity

import "github.com/pcornejov/juntalo/backend/internal/domain/apperr"

const minPasswordLength = 8

// ValidatePassword applies the MVP password rule: length only, no symbol/uppercase
// ceremony (Etapa 1: la experiencia debe ser simple).
func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return apperr.New("weak_password", "La contraseña debe tener al menos 8 caracteres")
	}
	return nil
}
