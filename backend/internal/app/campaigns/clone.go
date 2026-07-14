package campaigns

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

type CloneService struct {
	repo   app.CampaignRepository
	create *CreateService
}

func NewCloneService(repo app.CampaignRepository, create *CreateService) *CloneService {
	return &CloneService{repo: repo, create: create}
}

// Clone copies type_key, title (con sufijo "(copia)") y goal_amount de una
// campaña propia a una nueva campaña en borrador — sin fotos, sin
// participantes, sin fechas de inicio/término (Etapa 4: el organizador
// decide esos datos de nuevo al publicar la copia).
func (s *CloneService) Clone(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, error) {
	original, found, err := s.repo.GetByIDForOrg(ctx, id, orgID)
	if err != nil {
		return campaign.Campaign{}, err
	}
	if !found {
		return campaign.Campaign{}, ErrNotFound
	}

	return s.create.Create(ctx, CreateInput{
		OrganizationID: orgID,
		TypeKey:        original.TypeKey,
		Title:          original.Title + " (copia)",
		Description:    original.Description,
		GoalAmount:     original.GoalAmount,
	})
}
