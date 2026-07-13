package main

import (
	"context"
	"log"

	"github.com/pcornejov/juntalo/backend/internal/api"
	"github.com/pcornejov/juntalo/backend/internal/infra/config"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	app := api.NewServer(db)

	log.Printf("juntalo-api listening on :%s (env=%s)", cfg.Port, cfg.Env)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
