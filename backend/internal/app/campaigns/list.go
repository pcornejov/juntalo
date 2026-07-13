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
		totals, err := s.repo.GetTotals(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, CampaignWithTotals{Campaign: c, Totals: totals})
	}
	return out, nil
}
