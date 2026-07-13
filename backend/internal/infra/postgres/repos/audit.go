package repos

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type AuditRepo struct {
	q *sqlc.Queries
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{q: sqlc.New(pool)}
}

func (r *AuditRepo) Record(ctx context.Context, in app.RecordAuditInput) error {
	data, err := json.Marshal(in.Data)
	if err != nil {
		return fmt.Errorf("marshal audit data: %w", err)
	}

	if err := r.q.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		ActorUserID:    toPgUUID(&in.ActorUserID),
		OrganizationID: toPgUUID(&in.OrganizationID),
		Action:         in.Action,
		EntityType:     in.EntityType,
		EntityID:       in.EntityID,
		Data:           data,
	}); err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}
