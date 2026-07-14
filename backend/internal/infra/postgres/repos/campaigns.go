package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type CampaignRepo struct {
	q *sqlc.Queries
}

func NewCampaignRepo(pool *pgxpool.Pool) *CampaignRepo {
	return &CampaignRepo{q: sqlc.New(pool)}
}

func (r *CampaignRepo) Create(ctx context.Context, in app.CreateCampaignInput) (campaign.Campaign, error) {
	c, err := r.q.CreateCampaign(ctx, sqlc.CreateCampaignParams{
		OrganizationID:     in.OrganizationID,
		TypeKey:            string(in.TypeKey),
		Category:           string(in.Category),
		Title:              in.Title,
		Slug:               in.Slug,
		Description:        in.Description,
		GoalAmount:         toInt8(in.GoalAmount),
		StartsAt:           toTimestamptz(in.StartsAt),
		EndsAt:             toTimestamptz(in.EndsAt),
		PublishAt:          toTimestamptz(in.PublishAt),
		VideoUrl:           toPgText(in.VideoURL),
		RaffleUnitPrice:    toInt8(in.RaffleUnitPrice),
		RaffleTotalNumbers: toInt4(in.RaffleTotalNumbers),
	})
	if err != nil {
		return campaign.Campaign{}, fmt.Errorf("create campaign: %w", err)
	}
	return mapCampaign(c), nil
}

func (r *CampaignRepo) GetByID(ctx context.Context, id uuid.UUID) (campaign.Campaign, bool, error) {
	c, err := r.q.GetCampaignByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaign.Campaign{}, false, nil
		}
		return campaign.Campaign{}, false, fmt.Errorf("get campaign: %w", err)
	}
	return mapCampaign(c), true, nil
}

func (r *CampaignRepo) GetByIDForOrg(ctx context.Context, id, orgID uuid.UUID) (campaign.Campaign, bool, error) {
	c, err := r.q.GetCampaignByIDForOrg(ctx, sqlc.GetCampaignByIDForOrgParams{ID: id, OrganizationID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaign.Campaign{}, false, nil
		}
		return campaign.Campaign{}, false, fmt.Errorf("get campaign for org: %w", err)
	}
	return mapCampaign(c), true, nil
}

func (r *CampaignRepo) GetBySlug(ctx context.Context, slug string) (campaign.Campaign, bool, error) {
	c, err := r.q.GetCampaignBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaign.Campaign{}, false, nil
		}
		return campaign.Campaign{}, false, fmt.Errorf("get campaign by slug: %w", err)
	}
	return mapCampaign(c), true, nil
}

func (r *CampaignRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	exists, err := r.q.SlugExists(ctx, slug)
	if err != nil {
		return false, fmt.Errorf("slug exists: %w", err)
	}
	return exists, nil
}

func (r *CampaignRepo) ListByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]campaign.Campaign, error) {
	rows, err := r.q.ListCampaignsByOrg(ctx, sqlc.ListCampaignsByOrgParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}
	out := make([]campaign.Campaign, len(rows))
	for i, c := range rows {
		out[i] = mapCampaign(c)
	}
	return out, nil
}

func (r *CampaignRepo) ListPublic(ctx context.Context, search string, category campaign.Category, limit, offset int32) ([]campaign.Campaign, error) {
	rows, err := r.q.ListActiveCampaigns(ctx, sqlc.ListActiveCampaignsParams{
		Limit:    limit,
		Offset:   offset,
		Search:   pgtype.Text{String: search, Valid: search != ""},
		Category: pgtype.Text{String: string(category), Valid: category != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("list public campaigns: %w", err)
	}
	out := make([]campaign.Campaign, len(rows))
	for i, c := range rows {
		out[i] = mapCampaign(c)
	}
	return out, nil
}

func (r *CampaignRepo) ListPublicByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]campaign.Campaign, error) {
	rows, err := r.q.ListPublicCampaignsByOrg(ctx, sqlc.ListPublicCampaignsByOrgParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list public campaigns by org: %w", err)
	}
	out := make([]campaign.Campaign, len(rows))
	for i, c := range rows {
		out[i] = mapCampaign(c)
	}
	return out, nil
}

func (r *CampaignRepo) GetFeatured(ctx context.Context) (campaign.Campaign, bool, error) {
	c, err := r.q.GetFeaturedCampaign(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaign.Campaign{}, false, nil
		}
		return campaign.Campaign{}, false, fmt.Errorf("get featured campaign: %w", err)
	}
	return mapCampaign(c), true, nil
}

func (r *CampaignRepo) Update(ctx context.Context, in app.UpdateCampaignInput) (campaign.Campaign, error) {
	c, err := r.q.UpdateCampaign(ctx, sqlc.UpdateCampaignParams{
		ID:                  in.ID,
		Title:               in.Title,
		Description:         in.Description,
		GoalAmount:          toInt8(in.GoalAmount),
		StartsAt:            toTimestamptz(in.StartsAt),
		EndsAt:              toTimestamptz(in.EndsAt),
		CoverFileID:         toPgUUID(in.CoverFileID),
		PublishAt:           toTimestamptz(in.PublishAt),
		Category:            string(in.Category),
		VideoUrl:            toPgText(in.VideoURL),
		RaffleUnitPrice:     toInt8(in.RaffleUnitPrice),
		RaffleTotalNumbers:  toInt4(in.RaffleTotalNumbers),
		RaffleWinningNumber: toInt4(in.RaffleWinningNumber),
	})
	if err != nil {
		return campaign.Campaign{}, fmt.Errorf("update campaign: %w", err)
	}
	return mapCampaign(c), nil
}

func (r *CampaignRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status campaign.Status) (campaign.Campaign, error) {
	c, err := r.q.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{ID: id, Status: string(status)})
	if err != nil {
		return campaign.Campaign{}, fmt.Errorf("update campaign status: %w", err)
	}
	return mapCampaign(c), nil
}

func (r *CampaignRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.SoftDeleteCampaign(ctx, id); err != nil {
		return fmt.Errorf("soft delete campaign: %w", err)
	}
	return nil
}

func (r *CampaignRepo) PublishDueCampaigns(ctx context.Context) ([]campaign.Campaign, error) {
	rows, err := r.q.PublishDueCampaigns(ctx)
	if err != nil {
		return nil, fmt.Errorf("publish due campaigns: %w", err)
	}
	out := make([]campaign.Campaign, len(rows))
	for i, c := range rows {
		out[i] = mapCampaign(c)
	}
	return out, nil
}

func (r *CampaignRepo) GetTotals(ctx context.Context, id uuid.UUID) (campaign.Totals, error) {
	t, err := r.q.GetCampaignTotals(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaign.Totals{}, nil
		}
		return campaign.Totals{}, fmt.Errorf("get campaign totals: %w", err)
	}
	return campaign.Totals{
		RaisedGross:      money.CLP(t.RaisedGross),
		RaisedNetApprox:  money.CLP(t.RaisedNetApprox),
		ContributorCount: t.ContributorCount,
	}, nil
}

func (r *CampaignRepo) GetRaffleNumbersSold(ctx context.Context, campaignID uuid.UUID) (int64, error) {
	count, err := r.q.CountReservedRaffleNumbers(ctx, campaignID)
	if err != nil {
		return 0, fmt.Errorf("get raffle numbers sold: %w", err)
	}
	return count, nil
}

func toInt8(v *money.CLP) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(*v), Valid: true}
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toPgText(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}
