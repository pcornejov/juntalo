package campaigns

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

type UpdateService struct {
	repo app.CampaignRepository
}

func NewUpdateService(repo app.CampaignRepository) *UpdateService {
	return &UpdateService{repo: repo}
}

type UpdateInput struct {
	Title       string
	Description string
	GoalAmount  *money.CLP
	StartsAt    *time.Time
	EndsAt      *time.Time
	CoverFileID *uuid.UUID
	// PublishAt sigue el mismo patrón "mantener si no viene" que CoverFileID.
	// ClearPublishAt es el único camino para borrar una publicación programada
	// (un PublishAt nil por sí solo, con JSON omitiendo la clave, no debe
	// wipear silenciosamente la fecha en cada edición no relacionada).
	PublishAt      *time.Time
	ClearPublishAt bool
	Category       campaign.Category
	// VideoURL sigue el mismo patrón "mantener si no viene" que CoverFileID.
	// ClearVideoURL es el único camino para quitar un video ya asociado.
	VideoURL      *string
	ClearVideoURL bool
}

func (s *UpdateService) Update(ctx context.Context, id, orgID uuid.UUID, in UpdateInput) (campaign.Campaign, error) {
	existing, found, err := s.repo.GetByIDForOrg(ctx, id, orgID)
	if err != nil {
		return campaign.Campaign{}, err
	}
	if !found {
		return campaign.Campaign{}, ErrNotFound
	}

	if err := campaign.ValidateTitle(in.Title); err != nil {
		return campaign.Campaign{}, err
	}

	totals, err := s.repo.GetTotals(ctx, id)
	if err != nil {
		return campaign.Campaign{}, err
	}
	if err := campaign.ValidateGoalUpdate(in.GoalAmount, totals.RaisedGross); err != nil {
		return campaign.Campaign{}, err
	}

	coverFileID := existing.CoverFileID
	if in.CoverFileID != nil {
		coverFileID = in.CoverFileID
	}

	publishAt := existing.PublishAt
	if in.ClearPublishAt {
		publishAt = nil
	} else if in.PublishAt != nil {
		publishAt = in.PublishAt
	}
	if err := campaign.ValidatePublishAt(publishAt, existing.Status); err != nil {
		return campaign.Campaign{}, err
	}
	if err := campaign.ValidateCategory(in.Category); err != nil {
		return campaign.Campaign{}, err
	}
	category := existing.Category
	if in.Category != "" {
		category = in.Category
	}

	videoURL := existing.VideoURL
	if in.ClearVideoURL {
		videoURL = nil
	} else if in.VideoURL != nil {
		if err := campaign.ValidateVideoURL(*in.VideoURL); err != nil {
			return campaign.Campaign{}, err
		}
		videoURL = in.VideoURL
	}

	return s.repo.Update(ctx, app.UpdateCampaignInput{
		ID:          id,
		Title:       in.Title,
		Description: in.Description,
		GoalAmount:  in.GoalAmount,
		StartsAt:    in.StartsAt,
		EndsAt:      in.EndsAt,
		CoverFileID: coverFileID,
		PublishAt:   publishAt,
		Category:    category,
		VideoURL:    videoURL,
	})
}

// CancelSchedule clears a campaign's publish_at, leaving every other field
// untouched — un atajo sobre Update para el botón "cancelar publicación
// programada" del dashboard, que no debería tener que reenviar título,
// descripción, etc. solo para borrar una fecha.
func (s *UpdateService) CancelSchedule(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, error) {
	existing, found, err := s.repo.GetByIDForOrg(ctx, id, orgID)
	if err != nil {
		return campaign.Campaign{}, err
	}
	if !found {
		return campaign.Campaign{}, ErrNotFound
	}

	return s.repo.Update(ctx, app.UpdateCampaignInput{
		ID:          id,
		Title:       existing.Title,
		Description: existing.Description,
		GoalAmount:  existing.GoalAmount,
		StartsAt:    existing.StartsAt,
		EndsAt:      existing.EndsAt,
		CoverFileID: existing.CoverFileID,
		PublishAt:   nil,
		Category:    existing.Category,
		VideoURL:    existing.VideoURL,
	})
}
