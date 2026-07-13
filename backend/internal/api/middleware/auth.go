package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/app"
)

const userIDLocalsKey = "user_id"

// RequireAuth parses the Bearer access token and stores the user id in c.Locals
// for downstream handlers (Etapa 4 §1).
func RequireAuth(signer app.TokenSigner) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			return dto.WriteError(c, unauthorized())
		}

		userID, err := signer.Parse(token)
		if err != nil {
			return dto.WriteError(c, unauthorized())
		}

		c.Locals(userIDLocalsKey, userID)
		return c.Next()
	}
}

// UserID reads the authenticated user id set by RequireAuth.
func UserID(c *fiber.Ctx) uuid.UUID {
	id, _ := c.Locals(userIDLocalsKey).(uuid.UUID)
	return id
}
