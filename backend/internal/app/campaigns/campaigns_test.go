package campaigns

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

// fakeCampaignRepo implements app.CampaignRepository entirely in memory —
// exercises the use cases (slug uniqueness, transitions, org scoping)
// without touching Postgres.
type fakeCampaignRepo struct {
	byID  map[uuid.UUID]campaign.Campaign
	slugs map[string]bool
}

func newFakeCampaignRepo() *fakeCampaignRepo {
	return &fakeCampaignRepo{byID: map[uuid.UUID]campaign.Campaign{}, slugs: map[string]bool{}}
}

func (f *fakeCampaignRepo) Create(_ context.Context, in app.CreateCampaignInput) (campaign.Campaign, error) {
	c := campaign.Campaign{
		ID:             uuid.New(),
		OrganizationID: in.OrganizationID,
		TypeKey:        in.TypeKey,
		Title:          in.Title,
		Slug:           in.Slug,
		Description:    in.Description,
		GoalAmount:     in.GoalAmount,
		Status:         campaign.StatusDraft,
		StartsAt:       in.StartsAt,
		EndsAt:         in.EndsAt,
	}
	f.byID[c.ID] = c
	f.slugs[c.Slug] = true
	return c, nil
}

func (f *fakeCampaignRepo) GetByID(_ context.Context, id uuid.UUID) (campaign.Campaign, bool, error) {
	c, ok := f.byID[id]
	return c, ok, nil
}

func (f *fakeCampaignRepo) GetByIDForOrg(_ context.Context, id, orgID uuid.UUID) (campaign.Campaign, bool, error) {
	c, ok := f.byID[id]
	if !ok || c.OrganizationID != orgID {
		return campaign.Campaign{}, false, nil
	}
	return c, true, nil
}

func (f *fakeCampaignRepo) GetBySlug(_ context.Context, slug string) (campaign.Campaign, bool, error) {
	for _, c := range f.byID {
		if c.Slug == slug {
			return c, true, nil
		}
	}
	return campaign.Campaign{}, false, nil
}

func (f *fakeCampaignRepo) SlugExists(_ context.Context, slug string) (bool, error) {
	return f.slugs[slug], nil
}

func (f *fakeCampaignRepo) ListByOrg(_ context.Context, orgID uuid.UUID, limit, offset int32) ([]campaign.Campaign, error) {
	var out []campaign.Campaign
	for _, c := range f.byID {
		if c.OrganizationID == orgID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *fakeCampaignRepo) Update(_ context.Context, in app.UpdateCampaignInput) (campaign.Campaign, error) {
	c := f.byID[in.ID]
	c.Title = in.Title
	c.Description = in.Description
	c.GoalAmount = in.GoalAmount
	c.StartsAt = in.StartsAt
	c.EndsAt = in.EndsAt
	c.CoverFileID = in.CoverFileID
	f.byID[in.ID] = c
	return c, nil
}

func (f *fakeCampaignRepo) UpdateStatus(_ context.Context, id uuid.UUID, status campaign.Status) (campaign.Campaign, error) {
	c := f.byID[id]
	c.Status = status
	f.byID[id] = c
	return c, nil
}

func (f *fakeCampaignRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	delete(f.byID, id)
	return nil
}

func (f *fakeCampaignRepo) GetTotals(_ context.Context, id uuid.UUID) (campaign.Totals, error) {
	return campaign.Totals{}, nil
}

func (f *fakeCampaignRepo) PublishDueCampaigns(context.Context) ([]campaign.Campaign, error) {
	return nil, nil
}

func (f *fakeCampaignRepo) ListPublic(context.Context, string, int32, int32) ([]campaign.Campaign, error) {
	return nil, nil
}
