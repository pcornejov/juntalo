package repos

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type CampaignImageRepo struct {
	q *sqlc.Queries
}

func NewCampaignImageRepo(pool *pgxpool.Pool) *CampaignImageRepo {
	return &CampaignImageRepo{q: sqlc.New(pool)}
}

func (r *CampaignImageRepo) Add(ctx context.Context, campaignID, fileID uuid.UUID) (uuid.UUID, error) {
	// La posición se calcula en dos pasos (no una sola query con subselect
	// atómico) porque la carga de imágenes de un organizador no es
	// concurrente en la práctica (un formulario, un usuario) — se prioriza
	// la simplicidad del código generado sobre una garantía que aquí no
	// aporta valor real.
	pos, err := r.q.NextCampaignImagePosition(ctx, campaignID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("campaign images: next position: %w", err)
	}
	img, err := r.q.CreateCampaignImage(ctx, sqlc.CreateCampaignImageParams{
		CampaignID: campaignID,
		FileID:     fileID,
		Position:   pos,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("campaign images: create: %w", err)
	}
	return img.ID, nil
}

func (r *CampaignImageRepo) ListByCampaign(ctx context.Context, campaignID uuid.UUID) ([]app.CampaignImageRecord, error) {
	rows, err := r.q.ListCampaignImages(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaign images: list: %w", err)
	}
	out := make([]app.CampaignImageRecord, len(rows))
	for i, row := range rows {
		out[i] = app.CampaignImageRecord{ID: row.ID, StorageKey: row.StorageKey}
	}
	return out, nil
}

func (r *CampaignImageRepo) Delete(ctx context.Context, campaignID, imageID uuid.UUID) (bool, error) {
	affected, err := r.q.DeleteCampaignImage(ctx, sqlc.DeleteCampaignImageParams{
		ID:         imageID,
		CampaignID: campaignID,
	})
	if err != nil {
		return false, fmt.Errorf("campaign images: delete: %w", err)
	}
	return affected > 0, nil
}
