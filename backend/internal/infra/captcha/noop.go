package captcha

import "context"

// NoopVerifier siempre aprueba — usado cuando TURNSTILE_SECRET_KEY queda
// vacía (desarrollo local, tests, o un deploy que todavía no lo configuró),
// mismo patrón "vacío = off" que R2/Resend/Sentry/Webpay.
type NoopVerifier struct{}

func (NoopVerifier) Verify(ctx context.Context, token, remoteIP string) (bool, error) {
	return true, nil
}
