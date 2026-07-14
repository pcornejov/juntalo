package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type AdminRepo struct {
	q *sqlc.Queries
}

func NewAdminRepo(pool *pgxpool.Pool) *AdminRepo {
	return &AdminRepo{q: sqlc.New(pool)}
}

func (r *AdminRepo) ListUsers(ctx context.Context, limit, offset int32) ([]app.AdminUserRow, error) {
	rows, err := r.q.ListAdminUsers(ctx, sqlc.ListAdminUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list admin users: %w", err)
	}
	out := make([]app.AdminUserRow, len(rows))
	for i, row := range rows {
		out[i] = app.AdminUserRow{
			ID:               row.ID,
			Email:            row.Email,
			FullName:         row.FullName,
			EmailVerified:    row.EmailVerifiedAt.Valid,
			CreatedAt:        row.CreatedAt.Time,
			OrganizationID:   row.OrganizationID,
			OrganizationName: row.OrganizationName,
			CampaignCount:    row.CampaignCount,
		}
	}
	return out, nil
}

func (r *AdminRepo) ListCampaigns(ctx context.Context, limit, offset int32) ([]app.AdminCampaignRow, error) {
	rows, err := r.q.ListAdminCampaigns(ctx, sqlc.ListAdminCampaignsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list admin campaigns: %w", err)
	}
	out := make([]app.AdminCampaignRow, len(rows))
	for i, row := range rows {
		out[i] = app.AdminCampaignRow{
			ID:               row.ID,
			Title:            row.Title,
			Slug:             row.Slug,
			TypeKey:          campaign.TypeKey(row.TypeKey),
			Status:           campaign.Status(row.Status),
			Category:         campaign.Category(row.Category),
			CreatedAt:        row.CreatedAt.Time,
			OrganizationName: row.OrganizationName,
			OrganizerEmail:   row.OrganizerEmail,
			RaisedGross:      money.CLP(row.RaisedGross),
			ContributorCount: row.ContributorCount,
		}
	}
	return out, nil
}

func (r *AdminRepo) ListPayments(ctx context.Context, limit, offset int32) ([]app.AdminPaymentRow, error) {
	rows, err := r.q.ListAdminPayments(ctx, sqlc.ListAdminPaymentsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list admin payments: %w", err)
	}
	out := make([]app.AdminPaymentRow, len(rows))
	for i, row := range rows {
		p := app.AdminPaymentRow{
			ID:               row.ID,
			Status:           payment.Status(row.Status),
			Provider:         row.Provider,
			AmountGross:      money.CLP(row.AmountGross),
			AmountNet:        money.CLP(row.AmountNet),
			CommissionAmount: money.CLP(row.CommissionAmount),
			CreatedAt:        row.CreatedAt.Time,
			CampaignTitle:    row.CampaignTitle,
			CampaignSlug:     row.CampaignSlug,
			ContributorName:  row.ContributorName,
		}
		if row.ConfirmedAt.Valid {
			p.ConfirmedAt = &row.ConfirmedAt.Time
		}
		out[i] = p
	}
	return out, nil
}

func (r *AdminRepo) GetMetrics(ctx context.Context) (app.AdminMetrics, error) {
	row, err := r.q.GetAdminMetrics(ctx)
	if err != nil {
		return app.AdminMetrics{}, fmt.Errorf("get admin metrics: %w", err)
	}
	return app.AdminMetrics{
		TotalUsers:         row.TotalUsers,
		TotalCampaigns:     row.TotalCampaigns,
		ActiveCampaigns:    row.ActiveCampaigns,
		DraftCampaigns:     row.DraftCampaigns,
		FinishedCampaigns:  row.FinishedCampaigns,
		TotalContributions: row.TotalContributions,
		RaisedGross:        money.CLP(row.RaisedGross),
		RaisedNetApprox:    money.CLP(row.RaisedNetApprox),
		TotalCommission:    money.CLP(row.TotalCommission),
	}, nil
}
