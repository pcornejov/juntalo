// Package db embeds the SQL migrations and can apply them programmatically.
// Solo se usa cuando RUN_MIGRATIONS_ON_BOOT=true — hosting gratuito sin acceso
// a shell (p.ej. Render free tier) para probar el MVP. En el VPS real las
// migraciones siguen siendo un paso explícito de deploy (Etapa 5/6), nunca
// automático al arrancar el binario.
package db

import (
	"embed"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(databaseURL string) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate: source: %w", err)
	}

	// El driver pgx/v5 de golang-migrate se registra como "pgx5"; acepta
	// cualquier URL postgres://... siempre que el scheme diga pgx5.
	pgx5URL := "pgx5://" + strings.TrimPrefix(strings.TrimPrefix(databaseURL, "postgres://"), "postgresql://")

	m, err := migrate.NewWithSourceInstance("iofs", src, pgx5URL)
	if err != nil {
		return fmt.Errorf("migrate: init: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate: up: %w", err)
	}
	return nil
}
