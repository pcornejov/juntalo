// Package config loads typed configuration from environment variables, failing fast
// on startup if anything required is missing (Etapa 2 §2.5).
package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret   string `envconfig:"JWT_SECRET" required:"true"`
	Env         string `envconfig:"APP_ENV" default:"development"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
