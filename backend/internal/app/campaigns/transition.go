package campaigns

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

// TransitionService implements the state-machine actions of Etapa 4 §3
// (publish/pause/resume/finish) as explicit use cases, not a generic
// PATCH{status} — each one names the domain intent.
type TransitionService struct {
	repo app.CampaignRepository
}

func NewTransitionService(repo app.CampaignRepository) *TransitionService {
	return &TransitionService{repo: repo}
}

func (s *TransitionService) Publish(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, error) {
	return s.transition(ctx, id, orgID, campaign.StatusActive)
}

func (s *TransitionService) Pause(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, error) {
	return s.transition(ctx, id, orgID, campaign.StatusPaused)
}

func (s *TransitionService) Resume(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, error) {
	return s.transition(ctx, id, orgID, campaign.StatusActive)
}

func (s *TransitionService) Finish(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, error) {
	return s.transition(ctx, id, orgID, campaign.StatusFinished)
}

func (s *TransitionService) transition(ctx context.Context, id, orgID uuid.UUID, to campaign.Status) (campaign.Campaign, error) {
	c, found, err := s.repo.GetByIDForOrg(ctx, id, orgID)
	if err != nil {
		return campaign.Campaign{}, err
	}
	if !found {
		return campaign.Campaign{}, ErrNotFound
	}

	if err := campaign.ValidateTransition(c.Status, to); err != nil {
		return campaign.Campaign{}, err
	}

	return s.repo.UpdateStatus(ctx, id, to)
}
