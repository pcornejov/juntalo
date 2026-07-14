package repos

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

func mapUser(u sqlc.User) identity.User {
	return identity.User{
		ID:            u.ID,
		Email:         u.Email,
		FullName:      u.FullName,
		Status:        u.Status,
		EmailVerified: u.EmailVerifiedAt.Valid,
		CreatedAt:     u.CreatedAt.Time,
	}
}

func mapOrganization(o sqlc.Organization) identity.Organization {
	rate, _ := o.CommissionRate.Float64Value()
	return identity.Organization{
		ID:             o.ID,
		Name:           o.Name,
		Kind:           o.Kind,
		CommissionRate: rate.Float64,
	}
}

func mapContributor(c sqlc.Contributor) contribution.Contributor {
	return contribution.Contributor{
		ID:       c.ID,
		FullName: c.FullName,
		Email:    c.Email.String,
		Phone:    c.Phone.String,
	}
}

func mapRefreshToken(t sqlc.RefreshToken) identity.RefreshToken {
	rt := identity.RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt.Time,
	}
	if t.RevokedAt.Valid {
		rt.RevokedAt = &t.RevokedAt.Time
	}
	return rt
}

func mapCampaign(c sqlc.Campaign) campaign.Campaign {
	out := campaign.Campaign{
		ID:             c.ID,
		OrganizationID: c.OrganizationID,
		TypeKey:        campaign.TypeKey(c.TypeKey),
		Category:       campaign.Category(c.Category),
		Title:          c.Title,
		Slug:           c.Slug,
		Description:    c.Description,
		Status:         campaign.Status(c.Status),
		CreatedAt:      c.CreatedAt.Time,
		UpdatedAt:      c.UpdatedAt.Time,
	}
	if c.CoverFileID.Valid {
		id := uuid.UUID(c.CoverFileID.Bytes)
		out.CoverFileID = &id
	}
	if c.GoalAmount.Valid {
		amount := money.CLP(c.GoalAmount.Int64)
		out.GoalAmount = &amount
	}
	if c.StartsAt.Valid {
		out.StartsAt = &c.StartsAt.Time
	}
	if c.EndsAt.Valid {
		out.EndsAt = &c.EndsAt.Time
	}
	if c.PublishAt.Valid {
		out.PublishAt = &c.PublishAt.Time
	}
	return out
}

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func toNumeric(rate float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(strconv.FormatFloat(rate, 'f', 4, 64))
	return n
}

func mapContribution(c sqlc.Contribution) contribution.Contribution {
	return contribution.Contribution{
		ID:            c.ID,
		CampaignID:    c.CampaignID,
		ContributorID: c.ContributorID,
		Amount:        money.CLP(c.Amount),
		IsAnonymous:   c.IsAnonymous,
		Message:       c.Message.String,
		Status:        contribution.Status(c.Status),
		CreatedAt:     c.CreatedAt.Time,
	}
}

func mapPayment(p sqlc.Payment) payment.Payment {
	rate, _ := p.CommissionRateApplied.Float64Value()
	out := payment.Payment{
		ID:                    p.ID,
		ContributionID:        p.ContributionID,
		IdempotencyKey:        p.IdempotencyKey,
		Provider:              p.Provider,
		ProviderRef:           p.ProviderRef.String,
		Status:                payment.Status(p.Status),
		AmountGross:           money.CLP(p.AmountGross),
		CommissionRateApplied: rate.Float64,
		CommissionAmount:      money.CLP(p.CommissionAmount),
		AmountNet:             money.CLP(p.AmountNet),
		CreatedAt:             p.CreatedAt.Time,
	}
	if p.ConfirmedAt.Valid {
		out.ConfirmedAt = &p.ConfirmedAt.Time
	}
	if p.FailedAt.Valid {
		out.FailedAt = &p.FailedAt.Time
	}
	var snapshot map[string]string
	if len(p.PayeeSnapshot) > 0 {
		_ = json.Unmarshal(p.PayeeSnapshot, &snapshot)
	}
	out.PayeeSnapshot = snapshot
	return out
}
