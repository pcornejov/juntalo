// Package dashboard implements the organizer-facing reporting use cases:
// participantes y export CSV (Hito 4).
package dashboard

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
)

var ErrCampaignNotFound = apperr.New("campaign_not_found", "Campaña no encontrada")

const defaultParticipantsLimit = 100

type ParticipantsService struct {
	campaigns    app.CampaignRepository
	participants app.ParticipantRepository
}

func NewParticipantsService(campaigns app.CampaignRepository, participants app.ParticipantRepository) *ParticipantsService {
	return &ParticipantsService{campaigns: campaigns, participants: participants}
}

// List scopes to the organizer's own campaign — un uuid ajeno responde
// campaign_not_found, nunca un error de permisos (Etapa 4 §3).
func (s *ParticipantsService) List(ctx context.Context, campaignID, orgID uuid.UUID) ([]app.ParticipantRow, error) {
	if _, found, err := s.campaigns.GetByIDForOrg(ctx, campaignID, orgID); err != nil {
		return nil, err
	} else if !found {
		return nil, ErrCampaignNotFound
	}
	return s.participants.ListByCampaign(ctx, campaignID, defaultParticipantsLimit, 0)
}
