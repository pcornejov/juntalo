package middleware

import "github.com/pcornejov/juntalo/backend/internal/domain/apperr"

func unauthorized() error {
	return apperr.New("unauthorized", "Token inválido o ausente")
}

// notFound se usa para el gate de admin (RequireAdminUser): un usuario sin
// acceso al backoffice recibe 404, nunca un 403 — no confirma ni niega que
// el backoffice existe, mismo criterio que el resto de la app para recursos
// ajenos (Etapa 4 §1).
func notFound() error {
	return apperr.New("not_found", "No encontrado")
}
