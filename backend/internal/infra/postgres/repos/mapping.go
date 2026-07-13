package repos

import (
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
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
