package handlers

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/app"
)

// resolveGalleryImages devuelve las imágenes de una campaña en orden de
// posición, con su id (necesario para poder borrarlas desde el dashboard
// del organizador). Si aún no tiene ninguna, devuelve slice vacío — nunca
// nil, para que el frontend reciba siempre `images: []` sin null-check.
func resolveGalleryImages(ctx context.Context, images app.CampaignImageRepository, storage app.FileStorage, campaignID uuid.UUID) []dto.CampaignImageResponse {
	records, err := images.ListByCampaign(ctx, campaignID)
	if err != nil {
		return []dto.CampaignImageResponse{}
	}
	out := make([]dto.CampaignImageResponse, len(records))
	for i, r := range records {
		out[i] = dto.CampaignImageResponse{ID: r.ID.String(), URL: storage.PublicURL(r.StorageKey)}
	}
	return out
}

// resolveGalleryURLs es la variante para la página pública, que no necesita
// (ni debe exponer) los ids de imagen — solo las URLs a mostrar.
func resolveGalleryURLs(ctx context.Context, images app.CampaignImageRepository, storage app.FileStorage, campaignID uuid.UUID) []string {
	records, err := images.ListByCampaign(ctx, campaignID)
	if err != nil {
		return []string{}
	}
	urls := make([]string, len(records))
	for i, r := range records {
		urls[i] = storage.PublicURL(r.StorageKey)
	}
	return urls
}
