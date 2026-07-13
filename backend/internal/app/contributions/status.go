package contributions

import (
	"context"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
)

var ErrNotFound = apperr.New("contribution_not_found", "Aporte no encontrado")

type StatusService struct {
	contributions app.ContributionRepository
}

func NewStatusService(contributions app.ContributionRepository) *StatusService {
	return &StatusService{contributions: contributions}
}

// GetStatus is polled by the confirmation screen after a contribution starts
// (Etapa 4 §4). El id es el capability token: solo quien aportó lo tiene.
func (s *StatusService) GetStatus(ctx context.Context, id uuid.UUID) (contribution.Status, error) {
	c, found, err := s.contributions.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if !found {
		return "", ErrNotFound
	}
	return c.Status, nil
}
