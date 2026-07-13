package campaigns

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

var errNotDeletable = apperr.New("campaign_not_deletable", "Solo se pueden eliminar campañas en borrador")

type DeleteService struct {
	repo app.CampaignRepository
}

func NewDeleteService(repo app.CampaignRepository) *DeleteService {
	return &DeleteService{repo: repo}
}

// Delete only soft-deletes draft campaigns without payments (Etapa 4 §3):
// cualquier otro estado usa Finish.
func (s *DeleteService) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	c, found, err := s.repo.GetByIDForOrg(ctx, id, orgID)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	if c.Status != campaign.StatusDraft {
		return errNotDeletable
	}
	return s.repo.SoftDelete(ctx, id)
}
