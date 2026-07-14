package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/app"
)

// RequireAdminUser debe montarse después de RequireAuth (necesita user_id ya
// en locals). adminEmails es la lista blanca de emails con acceso al
// backoffice — comparación case-insensitive, sin distinguir "por qué no":
// token inválido, usuario inexistente o email fuera de la lista responden
// todos 404 por igual.
func RequireAdminUser(users app.UserRepository, adminEmails []string) fiber.Handler {
	allow := make(map[string]bool, len(adminEmails))
	for _, e := range adminEmails {
		e = strings.ToLower(strings.TrimSpace(e))
		if e != "" {
			allow[e] = true
		}
	}

	return func(c *fiber.Ctx) error {
		user, found, err := users.GetByID(c.Context(), UserID(c))
		if err != nil {
			return dto.WriteError(c, err)
		}
		if !found || !allow[strings.ToLower(user.Email)] {
			return dto.WriteError(c, notFound())
		}
		return c.Next()
	}
}
