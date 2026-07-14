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

	// ExposeResetLinks: todavía no hay envío de email real, así que
	// "olvidé mi contraseña" no tiene forma de llegarle al usuario salvo
	// que la API devuelva el link directo — solo aceptable en este deploy
	// de prueba. Debe quedar en false en cualquier despliegue real, donde
	// el link se enviaría por email y nunca por la respuesta HTTP.
	ExposeResetLinks bool `envconfig:"EXPOSE_RESET_LINKS" default:"false"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
