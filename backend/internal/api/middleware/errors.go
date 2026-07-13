package middleware

import "github.com/pcornejov/juntalo/backend/internal/domain/apperr"

func unauthorized() error {
	return apperr.New("unauthorized", "Token inválido o ausente")
}
