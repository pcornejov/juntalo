package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// ContributeLimiter caps public contribution attempts per IP (Etapa 4 §4:
// 10/min) — sliding window in-memory, suficiente para 1 VPS (Etapa 2 §2.5);
// Redis-ready cuando haya más de una instancia.
func ContributeLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
}
