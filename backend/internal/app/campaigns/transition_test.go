package campaigns

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

func createTestCampaign(t *testing.T, repo *fakeCampaignRepo, orgID uuid.UUID) campaign.Campaign {
	t.Helper()
	c, err := NewCreateService(repo).Create(context.Background(), CreateInput{
		OrganizationID: orgID,
		TypeKey:        campaign.TypeCollection,
		Title:          "Campaña de prueba",
	})
	if err != nil {
		t.Fatalf("setup create: %v", err)
	}
	return c
}

func TestTransitionService_PublishThenFinish(t *testing.T) {
	repo := newFakeCampaignRepo()
	orgID := uuid.New()
	c := createTestCampaign(t, repo, orgID)
	svc := NewTransitionService(repo)

	published, err := svc.Publish(context.Background(), c.ID, orgID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.Status != campaign.StatusActive {
		t.Errorf("got status %s, want active", published.Status)
	}

	finished, err := svc.Finish(context.Background(), c.ID, orgID)
	if err != nil {
		t.Fatalf("finish: %v", err)
	}
	if finished.Status != campaign.StatusFinished {
		t.Errorf("got status %s, want finished", finished.Status)
	}
}

func TestTransitionService_IllegalTransition(t *testing.T) {
	repo := newFakeCampaignRepo()
	orgID := uuid.New()
	c := createTestCampaign(t, repo, orgID)
	svc := NewTransitionService(repo)

	// draft -> paused no está permitido (solo draft -> active).
	_, err := svc.Pause(context.Background(), c.ID, orgID)
	if !apperr.Is(err, "invalid_status_transition") {
		t.Fatalf("expected invalid_status_transition, got %v", err)
	}
}

func TestTransitionService_WrongOrgReturnsNotFound(t *testing.T) {
	repo := newFakeCampaignRepo()
	orgID := uuid.New()
	c := createTestCampaign(t, repo, orgID)
	svc := NewTransitionService(repo)

	otherOrg := uuid.New()
	_, err := svc.Publish(context.Background(), c.ID, otherOrg)
	if !apperr.Is(err, "campaign_not_found") {
		t.Fatalf("expected campaign_not_found for a foreign org (no filtrar existencia), got %v", err)
	}
}

func TestDeleteService_OnlyDraftIsDeletable(t *testing.T) {
	repo := newFakeCampaignRepo()
	orgID := uuid.New()
	c := createTestCampaign(t, repo, orgID)

	transitionSvc := NewTransitionService(repo)
	if _, err := transitionSvc.Publish(context.Background(), c.ID, orgID); err != nil {
		t.Fatalf("publish: %v", err)
	}

	deleteSvc := NewDeleteService(repo)
	err := deleteSvc.Delete(context.Background(), c.ID, orgID)
	if !apperr.Is(err, "campaign_not_deletable") {
		t.Fatalf("expected campaign_not_deletable for an active campaign, got %v", err)
	}
}
