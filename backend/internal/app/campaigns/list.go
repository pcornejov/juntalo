package campaigns

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

type ListService struct {
	repo app.CampaignRepository
}

func NewListService(repo app.CampaignRepository) *ListService {
	return &ListService{repo: repo}
}

type CampaignWithTotals struct {
	Campaign campaign.Campaign
	Totals   campaign.Totals
	// RaffleNumbersSold: solo se completa para TypeRaffle — números
	// reservados (pending) o vendidos (confirmed). El disponible se calcula
	// como RaffleTotalNumbers - RaffleNumbersSold.
	RaffleNumbersSold *int64
}

func (s *ListService) withTotals(ctx context.Context, c campaign.Campaign) (CampaignWithTotals, error) {
	totals, err := s.repo.GetTotals(ctx, c.ID)
	if err != nil {
		return CampaignWithTotals{}, err
	}
	out := CampaignWithTotals{Campaign: c, Totals: totals}
	if c.TypeKey == campaign.TypeRaffle {
		sold, err := s.repo.GetRaffleNumbersSold(ctx, c.ID)
		if err != nil {
			return CampaignWithTotals{}, err
		}
		out.RaffleNumbersSold = &sold
	}
	return out, nil
}

// List uses offset pagination: a single organizer has at most a handful of
// campaigns in the MVP, so the cursor-based scheme documented in Etapa 4 §1
// is deferred here — it would add complexity this list doesn't need yet.
func (s *ListService) List(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]CampaignWithTotals, error) {
	items, err := s.repo.ListByOrg(ctx, orgID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]CampaignWithTotals, 0, len(items))
	for _, c := range items {
		cwt, err := s.withTotals(ctx, c)
		if err != nil {
			return nil, err
		}
		out = append(out, cwt)
	}
	return out, nil
}

// ListPublic backs la sección pública de "Campañas activas" (Etapa 4): a
// diferencia de List, no filtra por organización — cualquier visitante debe
// poder explorar campañas de cualquier organizador para aportar.
func (s *ListService) ListPublic(ctx context.Context, search string, category campaign.Category, limit, offset int32) ([]CampaignWithTotals, error) {
	items, err := s.repo.ListPublic(ctx, search, category, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]CampaignWithTotals, 0, len(items))
	for _, c := range items {
		cwt, err := s.withTotals(ctx, c)
		if err != nil {
			return nil, err
		}
		out = append(out, cwt)
	}
	return out, nil
}

// ListPublicByOrg backs la página pública persistente del organizador
// (/org/:slug): igual que ListPublic pero acotado a una sola organización.
func (s *ListService) ListPublicByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]CampaignWithTotals, error) {
	items, err := s.repo.ListPublicByOrg(ctx, orgID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]CampaignWithTotals, 0, len(items))
	for _, c := range items {
		cwt, err := s.withTotals(ctx, c)
		if err != nil {
			return nil, err
		}
		out = append(out, cwt)
	}
	return out, nil
}

// GetFeatured backs la tarjeta "campaña destacada" en la sección pública
// (inspirado en Vaki): la campaña activa con el aporte confirmado más
// reciente. found=false si ninguna todavía tiene aportes confirmados.
func (s *ListService) GetFeatured(ctx context.Context) (CampaignWithTotals, bool, error) {
	c, found, err := s.repo.GetFeatured(ctx)
	if err != nil || !found {
		return CampaignWithTotals{}, false, err
	}
	cwt, err := s.withTotals(ctx, c)
	if err != nil {
		return CampaignWithTotals{}, false, err
	}
	return cwt, true, nil
}
