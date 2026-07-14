package main

import (
	"context"
	"log"
	"time"

	"github.com/getsentry/sentry-go"

	juntdb "github.com/pcornejov/juntalo/backend/db"
	"github.com/pcornejov/juntalo/backend/internal/api"
	"github.com/pcornejov/juntalo/backend/internal/infra/config"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// DSN vacío deja sentry-go en modo no-op: Init nunca falla ni bloquea el
	// arranque por esto (Config.SentryDSN).
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.Env,
		TracesSampleRate: 0, // solo error tracking por ahora, no performance tracing
	}); err != nil {
		log.Printf("sentry init: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	if cfg.RunMigrationsOnBoot {
		log.Println("running migrations (RUN_MIGRATIONS_ON_BOOT=true)...")
		if err := juntdb.RunMigrations(cfg.DatabaseURL); err != nil {
			log.Fatalf("migrations: %v", err)
		}
	}

	ctx := context.Background()
	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	app := api.NewServer(db, api.Config{
		JWTSecret:          cfg.JWTSecret,
		IsProd:             cfg.Env == "production",
		StorageDir:         cfg.StorageDir,
		StorageURL:         cfg.StorageURL,
		FrontendURL:        cfg.FrontendURL,
		SelfURL:            cfg.SelfURL,
		MockWebhookSecret:  cfg.MockWebhookSecret,
		MockPaymentMode:    cfg.MockPaymentMode,
		ExposeResetLinks:   cfg.ExposeResetLinks,
		ResendAPIKey:       cfg.ResendAPIKey,
		EmailFrom:          cfg.EmailFrom,
		R2AccountID:        cfg.R2AccountID,
		R2AccessKeyID:      cfg.R2AccessKeyID,
		R2SecretAccessKey:  cfg.R2SecretAccessKey,
		R2Bucket:           cfg.R2Bucket,
		R2PublicURL:        cfg.R2PublicURL,
		WebpayCommerceCode: cfg.WebpayCommerceCode,
		WebpayAPIKey:       cfg.WebpayAPIKey,
		WebpayEnvironment:  cfg.WebpayEnvironment,
		AdminEmails:        cfg.AdminEmails,
	})

	log.Printf("juntalo-api listening on :%s (env=%s)", cfg.Port, cfg.Env)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
