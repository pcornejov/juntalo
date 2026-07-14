package dashboard

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
)

type fakeCampaignRepo struct {
	byID map[uuid.UUID]campaign.Campaign
}

func (f *fakeCampaignRepo) Create(context.Context, app.CreateCampaignInput) (campaign.Campaign, error) {
	return campaign.Campaign{}, nil
}
func (f *fakeCampaignRepo) GetByID(context.Context, uuid.UUID) (campaign.Campaign, bool, error) {
	return campaign.Campaign{}, false, nil
}
func (f *fakeCampaignRepo) GetByIDForOrg(_ context.Context, id, orgID uuid.UUID) (campaign.Campaign, bool, error) {
	c, ok := f.byID[id]
	if !ok || c.OrganizationID != orgID {
		return campaign.Campaign{}, false, nil
	}
	return c, true, nil
}
func (f *fakeCampaignRepo) GetBySlug(context.Context, string) (campaign.Campaign, bool, error) {
	return campaign.Campaign{}, false, nil
}
func (f *fakeCampaignRepo) SlugExists(context.Context, string) (bool, error) { return false, nil }
func (f *fakeCampaignRepo) ListByOrg(context.Context, uuid.UUID, int32, int32) ([]campaign.Campaign, error) {
	return nil, nil
}
func (f *fakeCampaignRepo) Update(context.Context, app.UpdateCampaignInput) (campaign.Campaign, error) {
	return campaign.Campaign{}, nil
}
func (f *fakeCampaignRepo) UpdateStatus(context.Context, uuid.UUID, campaign.Status) (campaign.Campaign, error) {
	return campaign.Campaign{}, nil
}
func (f *fakeCampaignRepo) SoftDelete(context.Context, uuid.UUID) error { return nil }
func (f *fakeCampaignRepo) GetTotals(context.Context, uuid.UUID) (campaign.Totals, error) {
	return campaign.Totals{}, nil
}
func (f *fakeCampaignRepo) PublishDueCampaigns(context.Context) ([]campaign.Campaign, error) {
	return nil, nil
}

func (f *fakeCampaignRepo) ListPublic(context.Context, string, int32, int32) ([]campaign.Campaign, error) {
	return nil, nil
}

type fakeParticipantRepo struct {
	rows []app.ParticipantRow
}

func (f *fakeParticipantRepo) ListByCampaign(context.Context, uuid.UUID, int32, int32) ([]app.ParticipantRow, error) {
	return f.rows, nil
}

func (f *fakeParticipantRepo) ListByCampaignFiltered(_ context.Context, _ uuid.UUID, search, status string, _, _ int32) ([]app.ParticipantRow, error) {
	var out []app.ParticipantRow
	for _, r := range f.rows {
		if status != "" && string(r.Status) != status {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(r.FullName), strings.ToLower(search)) {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func TestParticipantsService_ForeignOrgReturnsNotFound(t *testing.T) {
	campaignID, ownerOrg, otherOrg := uuid.New(), uuid.New(), uuid.New()
	campaigns := &fakeCampaignRepo{byID: map[uuid.UUID]campaign.Campaign{
		campaignID: {ID: campaignID, OrganizationID: ownerOrg},
	}}
	svc := NewParticipantsService(campaigns, &fakeParticipantRepo{})

	_, err := svc.List(context.Background(), campaignID, otherOrg, "", "", 0, 0)
	if !apperr.Is(err, "campaign_not_found") {
		t.Fatalf("expected campaign_not_found for a foreign org (no filtrar existencia), got %v", err)
	}
}

func TestParticipantsService_FiltersBySearchAndStatus(t *testing.T) {
	campaignID, orgID := uuid.New(), uuid.New()
	campaigns := &fakeCampaignRepo{byID: map[uuid.UUID]campaign.Campaign{
		campaignID: {ID: campaignID, OrganizationID: orgID},
	}}
	participants := &fakeParticipantRepo{rows: []app.ParticipantRow{
		{FullName: "Zoe Findable", Status: contribution.StatusConfirmed},
		{FullName: "Ana Pérez", Status: contribution.StatusRefunded},
	}}
	svc := NewParticipantsService(campaigns, participants)

	got, err := svc.List(context.Background(), campaignID, orgID, "zoe", "", 100, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].FullName != "Zoe Findable" {
		t.Fatalf("expected search to find Zoe Findable, got %+v", got)
	}

	got, err = svc.List(context.Background(), campaignID, orgID, "", "refunded", 100, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].FullName != "Ana Pérez" {
		t.Fatalf("expected status filter to find Ana Pérez, got %+v", got)
	}
}

func TestExportCSVService_WriteCSV(t *testing.T) {
	campaignID, orgID := uuid.New(), uuid.New()
	campaigns := &fakeCampaignRepo{byID: map[uuid.UUID]campaign.Campaign{
		campaignID: {ID: campaignID, OrganizationID: orgID},
	}}
	participants := &fakeParticipantRepo{rows: []app.ParticipantRow{
		{
			FullName: "Ana Pérez", Email: "ana@example.com", Amount: 10_000,
			RefundedAmount: 2_000, IsAnonymous: false, Status: contribution.StatusConfirmed,
		},
		{
			FullName: "Anónimo", Amount: 5_000, IsAnonymous: true, Status: contribution.StatusConfirmed,
		},
	}}
	svc := NewExportCSVService(campaigns, participants)

	var buf bytes.Buffer
	if err := svc.WriteCSV(context.Background(), campaignID, orgID, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "fecha,nombre,email,telefono,monto,estado,anonimo,monto_reembolsado") {
		t.Errorf("missing expected header, got: %s", out)
	}
	if !strings.Contains(out, "Ana Pérez") || !strings.Contains(out, "10000") || !strings.Contains(out, "2000") {
		t.Errorf("missing expected row data, got: %s", out)
	}
	if !strings.Contains(out, ",si,") {
		t.Errorf("expected the anonymous row to be marked 'si', got: %s", out)
	}
}

func TestExportCSVService_ForeignOrgReturnsNotFound(t *testing.T) {
	campaignID, ownerOrg, otherOrg := uuid.New(), uuid.New(), uuid.New()
	campaigns := &fakeCampaignRepo{byID: map[uuid.UUID]campaign.Campaign{
		campaignID: {ID: campaignID, OrganizationID: ownerOrg},
	}}
	svc := NewExportCSVService(campaigns, &fakeParticipantRepo{})

	var buf bytes.Buffer
	err := svc.WriteCSV(context.Background(), campaignID, otherOrg, &buf)
	if !apperr.Is(err, "campaign_not_found") {
		t.Fatalf("expected campaign_not_found, got %v", err)
	}
}
