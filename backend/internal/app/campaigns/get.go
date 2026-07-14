package campaigns

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

var ErrNotFound = apperr.New("campaign_not_found", "Campaña no encontrada")

type GetService struct {
	repo app.CampaignRepository
}

func NewGetService(repo app.CampaignRepository) *GetService {
	return &GetService{repo: repo}
}

// GetForOrg returns the campaign only if it belongs to orgID — a uuid ajeno
// responde 404, nunca 403 (Etapa 4 §3: no filtrar existencia).
func (s *GetService) GetForOrg(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, campaign.Totals, error) {
	c, found, err := s.repo.GetByIDForOrg(ctx, id, orgID)
	if err != nil {
		return campaign.Campaign{}, campaign.Totals{}, err
	}
	if !found {
		return campaign.Campaign{}, campaign.Totals{}, ErrNotFound
	}

	totals, err := s.repo.GetTotals(ctx, id)
	if err != nil {
		return campaign.Campaign{}, campaign.Totals{}, err
	}
	return c, totals, nil
}

// GetPublicBySlug returns a campaign only if it's in a publicly visible status
// (Etapa 4 §4): draft y suspended responden 404, no filtran su existencia.
func (s *GetService) GetPublicBySlug(ctx context.Context, slug string) (campaign.Campaign, campaign.Totals, error) {
	c, found, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return campaign.Campaign{}, campaign.Totals{}, err
	}
	if !found || !publicStatuses[c.Status] {
		return campaign.Campaign{}, campaign.Totals{}, ErrNotFound
	}

	totals, err := s.repo.GetTotals(ctx, c.ID)
	if err != nil {
		return campaign.Campaign{}, campaign.Totals{}, err
	}
	return c, totals, nil
}

// GetRaffleNumbersSold expone el conteo de números reservados/vendidos para
// una campaña de tipo "raffle" — usado por los handlers que muestran
// disponibilidad (dashboard del organizador y página pública).
func (s *GetService) GetRaffleNumbersSold(ctx context.Context, campaignID uuid.UUID) (int64, error) {
	return s.repo.GetRaffleNumbersSold(ctx, campaignID)
}

var publicStatuses = map[campaign.Status]bool{
	campaign.StatusActive:   true,
	campaign.StatusPaused:   true,
	campaign.StatusFinished: true,
}
