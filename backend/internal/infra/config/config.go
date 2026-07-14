// Package config loads typed configuration from environment variables, failing fast
// on startup if anything required is missing (Etapa 2 §2.5).
package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret   string `envconfig:"JWT_SECRET" required:"true"`
	Env         string `envconfig:"APP_ENV" default:"development"`
	StorageDir  string `envconfig:"STORAGE_DIR" default:"./data/files"`
	StorageURL  string `envconfig:"STORAGE_URL" default:"http://localhost:8080/files"`
	FrontendURL string `envconfig:"FRONTEND_URL" default:"http://localhost:5173"`

	// Mock payment provider (Hito 3) — reemplazado por pasarelas reales post-MVP.
	SelfURL           string `envconfig:"SELF_URL" default:"http://localhost:8080"`
	MockWebhookSecret string `envconfig:"MOCK_WEBHOOK_SECRET" default:"dev-mock-secret"`
	MockPaymentMode   string `envconfig:"MOCK_PAYMENT_MODE" default:"deferred"`

	// RunMigrationsOnBoot solo se activa en hosting gratuito sin shell (Render);
	// el VPS real las corre como paso explícito de deploy (Etapa 5/6).
	RunMigrationsOnBoot bool `envconfig:"RUN_MIGRATIONS_ON_BOOT" default:"false"`

	// ExposeResetLinks: fallback para probar "olvidé mi contraseña" sin
	// RESEND_API_KEY configurada — solo aceptable en deploys de prueba, ver
	// email.go. Debe quedar en false en cualquier despliegue real.
	ExposeResetLinks bool `envconfig:"EXPOSE_RESET_LINKS" default:"false"`

	// Email transaccional (Resend). Si RESEND_API_KEY queda vacía, el envío
	// cae a un no-op que solo loguea — no rompe el flujo, simplemente no
	// llegan los emails (ver internal/infra/email).
	ResendAPIKey string `envconfig:"RESEND_API_KEY" default:""`
	EmailFrom    string `envconfig:"EMAIL_FROM" default:"onboarding@resend.dev"`

	// Monitoreo de errores (Sentry). DSN vacío = sentry-go queda en modo
	// no-op (no manda nada, no falla) — mismo patrón que RESEND_API_KEY.
	SentryDSN string `envconfig:"SENTRY_DSN" default:""`

	// Storage persistente (Cloudflare R2, S3-compatible). Si R2AccountID
	// queda vacío, el server cae a storage/local sobre StorageDir — sirve
	// para desarrollo local, pero en Render StorageDir vive en disco efímero
	// (se borra en cada redeploy), así que producción SIEMPRE debe traer
	// estas cuatro variables seteadas (ver render.yaml, todas sync: false).
	R2AccountID       string `envconfig:"R2_ACCOUNT_ID" default:""`
	R2AccessKeyID     string `envconfig:"R2_ACCESS_KEY_ID" default:""`
	R2SecretAccessKey string `envconfig:"R2_SECRET_ACCESS_KEY" default:""`
	R2Bucket          string `envconfig:"R2_BUCKET" default:""`
	// R2PublicURL: dominio público del bucket (r2.dev o un dominio custom
	// conectado) — no es el endpoint de la API S3, ese se arma con
	// R2AccountID (ver internal/infra/storage/r2).
	R2PublicURL string `envconfig:"R2_PUBLIC_URL" default:""`

	// Pasarela de pago real (Webpay Plus / Transbank). WebpayCommerceCode
	// vacío = cae a MockPaymentProvider — mismo patrón que R2AccountID.
	// Para "integration" el comercio y api key son públicos (los mismos
	// para todos los desarrolladores, ver render.yaml); para "production"
	// hay que reemplazarlos por los del comercio afiliado real de Transbank.
	WebpayCommerceCode string `envconfig:"WEBPAY_COMMERCE_CODE" default:""`
	WebpayAPIKey       string `envconfig:"WEBPAY_API_KEY" default:""`
	WebpayEnvironment  string `envconfig:"WEBPAY_ENVIRONMENT" default:"integration"`

	// Backoffice del operador de la plataforma. Vacío = nadie tiene acceso —
	// mismo patrón "vacío = off" que R2/Resend/Sentry/Webpay. Lista separada
	// por comas de emails con acceso (ver middleware.RequireAdminUser).
	AdminEmails string `envconfig:"ADMIN_EMAILS" default:""`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
