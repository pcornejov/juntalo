// Package organizations implements el perfil autoadministrado del
// organizador — hoy solo los datos de transferencia para la liquidación
// manual del backoffice.
package organizations

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

var errPayoutInfoIncomplete = apperr.New("validation_failed", "Completa todos los datos de transferencia")

type Service struct {
	repo app.OrganizationRepository
}

func NewService(repo app.OrganizationRepository) *Service {
	return &Service{repo: repo}
}

// UpdatePayoutInfo requires every field non-empty — datos bancarios a
// medias no sirven para transferir (Etapa post-MVP: liquidación manual).
func (s *Service) UpdatePayoutInfo(ctx context.Context, orgID uuid.UUID, in app.UpdatePayoutInfoInput) (identity.Organization, error) {
	fields := []string{in.Rut, in.PayoutBank, in.PayoutAccountType, in.PayoutAccountNumber, in.PayoutHolderName}
	for _, f := range fields {
		if strings.TrimSpace(f) == "" {
			return identity.Organization{}, errPayoutInfoIncomplete
		}
	}
	return s.repo.UpdatePayoutInfo(ctx, orgID, in)
}
