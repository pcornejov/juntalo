package campaigns

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

func TestCreateService_Create(t *testing.T) {
	repo := newFakeCampaignRepo()
	svc := NewCreateService(repo)
	orgID := uuid.New()

	c, err := svc.Create(context.Background(), CreateInput{
		OrganizationID: orgID,
		TypeKey:        campaign.TypeCollection,
		Title:          "Ayuda para el viaje de estudios",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Slug != "ayuda-para-el-viaje-de-estudios" {
		t.Errorf("got slug %q", c.Slug)
	}
	if c.Status != campaign.StatusDraft {
		t.Errorf("new campaign should start as draft, got %s", c.Status)
	}
}

func TestCreateService_DisabledType(t *testing.T) {
	repo := newFakeCampaignRepo()
	svc := NewCreateService(repo)

	_, err := svc.Create(context.Background(), CreateInput{
		OrganizationID: uuid.New(),
		TypeKey:        campaign.TypeRaffle,
		Title:          "Rifa de prueba",
	})
	if !apperr.Is(err, "campaign_type_disabled") {
		t.Fatalf("expected campaign_type_disabled, got %v", err)
	}
}

func TestCreateService_SlugCollisionGetsSuffix(t *testing.T) {
	repo := newFakeCampaignRepo()
	svc := NewCreateService(repo)
	orgID := uuid.New()

	first, err := svc.Create(context.Background(), CreateInput{OrganizationID: orgID, TypeKey: campaign.TypeCollection, Title: "Ayuda"})
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	second, err := svc.Create(context.Background(), CreateInput{OrganizationID: orgID, TypeKey: campaign.TypeCollection, Title: "Ayuda"})
	if err != nil {
		t.Fatalf("second create: %v", err)
	}

	if first.Slug == second.Slug {
		t.Errorf("expected different slugs for colliding titles, got the same: %q", first.Slug)
	}
}
