package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

type FileRepo struct {
	q *sqlc.Queries
}

func NewFileRepo(pool *pgxpool.Pool) *FileRepo {
	return &FileRepo{q: sqlc.New(pool)}
}

func (r *FileRepo) Create(ctx context.Context, in app.CreateFileInput) (uuid.UUID, error) {
	f, err := r.q.CreateFile(ctx, sqlc.CreateFileParams{
		OrganizationID: in.OrganizationID,
		Kind:           in.Kind,
		StorageKey:     in.StorageKey,
		MimeType:       in.MimeType,
		SizeBytes:      in.SizeBytes,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create file: %w", err)
	}
	return f.ID, nil
}

func (r *FileRepo) GetByID(ctx context.Context, id uuid.UUID) (app.FileRecord, bool, error) {
	f, err := r.q.GetFileByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.FileRecord{}, false, nil
		}
		return app.FileRecord{}, false, fmt.Errorf("get file: %w", err)
	}
	return app.FileRecord{ID: f.ID, StorageKey: f.StorageKey, MimeType: f.MimeType}, true, nil
}
