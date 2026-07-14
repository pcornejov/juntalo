package campaigns

import (
	"context"
	"log"
	"time"

	"github.com/pcornejov/juntalo/backend/internal/app"
)

// RunPublishScheduler polls every interval for draft campaigns whose
// publish_at ya venció y las publica — el "auto-publicar en fecha
// programada" corre dentro del mismo proceso del API (sin infra de cron
// aparte, consistente con el hosting gratuito de este MVP). Bloquea hasta
// que ctx se cancela; se lanza en su propia goroutine desde NewServer.
func RunPublishScheduler(ctx context.Context, repo app.CampaignRepository, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			published, err := repo.PublishDueCampaigns(ctx)
			if err != nil {
				log.Printf("publish scheduler: %v", err)
				continue
			}
			for _, c := range published {
				log.Printf("publish scheduler: auto-published campaign %s (%s)", c.ID, c.Slug)
			}
		}
	}
}
