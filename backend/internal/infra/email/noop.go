package email

import (
	"context"
	"log"
)

// NoopSender se usa cuando no hay RESEND_API_KEY configurada: no envía nada,
// solo deja rastro en el log — así un ambiente sin la key configurada no
// rompe el flujo (p.ej. registro), simplemente no llegan los emails.
type NoopSender struct{}

func NewNoopSender() *NoopSender { return &NoopSender{} }

func (s *NoopSender) Send(_ context.Context, to, subject, _ string) error {
	log.Printf("email (noop, RESEND_API_KEY no configurada): to=%s subject=%q", to, subject)
	return nil
}
