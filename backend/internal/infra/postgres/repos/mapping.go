package repos

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

func mapUser(u sqlc.User) identity.User {
	return identity.User{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Time,
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
	return out
}

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
