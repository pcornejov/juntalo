package email

import (
	"context"
	"log"
	"regexp"
)

// NoopSender se usa cuando no hay RESEND_API_KEY configurada: no envía nada,
// solo deja rastro en el log — así un ambiente sin la key configurada no
// rompe el flujo (p.ej. registro), simplemente no llegan los emails. Loguea
// el primer link del cuerpo (si hay) para poder probar flujos de un solo
// click — verificar email, resetear contraseña — sin bandeja de correo real,
// en el mismo espíritu que Config.ExposeResetLinks.
type NoopSender struct{}

func NewNoopSender() *NoopSender { return &NoopSender{} }

var firstHref = regexp.MustCompile(`href="([^"]+)"`)

func (s *NoopSender) Send(_ context.Context, to, subject, htmlBody string) error {
	link := ""
	if m := firstHref.FindStringSubmatch(htmlBody); m != nil {
		link = m[1]
	}
	log.Printf("email (noop, RESEND_API_KEY no configurada): to=%s subject=%q link=%s", to, subject, link)
	return nil
}
