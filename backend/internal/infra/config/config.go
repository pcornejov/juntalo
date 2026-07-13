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
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
