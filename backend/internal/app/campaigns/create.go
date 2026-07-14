package campaigns

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

type CreateService struct {
	repo app.CampaignRepository
}

func NewCreateService(repo app.CampaignRepository) *CreateService {
	return &CreateService{repo: repo}
}

type CreateInput struct {
	OrganizationID uuid.UUID
	TypeKey        campaign.TypeKey
	Category       campaign.Category
	Title          string
	Description    string
	GoalAmount     *money.CLP
	StartsAt       *time.Time
	EndsAt         *time.Time
	PublishAt      *time.Time
}

// Create validates the type + title, generates a server-side unique slug, and
// persists the campaign in draft (Etapa 4 §3: el slug lo genera el servidor).
func (s *CreateService) Create(ctx context.Context, in CreateInput) (campaign.Campaign, error) {
	if err := campaign.ValidateType(in.TypeKey); err != nil {
		return campaign.Campaign{}, err
	}
	if err := campaign.ValidateTitle(in.Title); err != nil {
		return campaign.Campaign{}, err
	}
	if err := campaign.ValidatePublishAt(in.PublishAt, campaign.StatusDraft); err != nil {
		return campaign.Campaign{}, err
	}
	if err := campaign.ValidateCategory(in.Category); err != nil {
		return campaign.Campaign{}, err
	}
	category := in.Category
	if category == "" {
		category = campaign.CategoryOtro
	}

	slug, err := s.uniqueSlug(ctx, campaign.Slugify(in.Title))
	if err != nil {
		return campaign.Campaign{}, err
	}

	return s.repo.Create(ctx, app.CreateCampaignInput{
		OrganizationID: in.OrganizationID,
		TypeKey:        in.TypeKey,
		Category:       category,
		Title:          strings.TrimSpace(in.Title),
		Slug:           slug,
		Description:    in.Description,
		GoalAmount:     in.GoalAmount,
		StartsAt:       in.StartsAt,
		EndsAt:         in.EndsAt,
		PublishAt:      in.PublishAt,
	})
}

const maxSlugAttempts = 5

func (s *CreateService) uniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 0; i < maxSlugAttempts; i++ {
		exists, err := s.repo.SlugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = campaign.WithSuffix(base, uuid.New().String()[:4])
	}
	return "", errors.New("campaigns: could not generate a unique slug")
}
