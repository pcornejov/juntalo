package dashboard

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"

	"github.com/google/uuid"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

const csvExportLimit = 10_000 // suficiente para el volumen de un organizador en el MVP

type ExportCSVService struct {
	campaigns    app.CampaignRepository
	participants app.ParticipantRepository
}

func NewExportCSVService(campaigns app.CampaignRepository, participants app.ParticipantRepository) *ExportCSVService {
	return &ExportCSVService{campaigns: campaigns, participants: participants}
}

// WriteCSV streams participantes a w. Columnas: fecha, nombre, email,
// teléfono, monto, estado, anónimo, monto_reembolsado (Etapa 4 §3) — el
// monto ya resta lo reembolsado en el reporting (Etapa 1 riesgo 10).
func (s *ExportCSVService) WriteCSV(ctx context.Context, campaignID, orgID uuid.UUID, w io.Writer) error {
	if _, found, err := s.campaigns.GetByIDForOrg(ctx, campaignID, orgID); err != nil {
		return err
	} else if !found {
		return ErrCampaignNotFound
	}

	rows, err := s.participants.ListByCampaign(ctx, campaignID, csvExportLimit, 0)
	if err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	header := []string{"fecha", "nombre", "email", "telefono", "monto", "estado", "anonimo", "monto_reembolsado", "mensaje"}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		record := []string{
			r.CreatedAt.Format("2006-01-02 15:04:05"),
			r.FullName,
			r.Email,
			r.Phone,
			formatCLPPlain(r.Amount),
			string(r.Status),
			formatBool(r.IsAnonymous),
			formatCLPPlain(r.RefundedAmount),
			r.Message,
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

func formatBool(b bool) string {
	if b {
		return "si"
	}
	return "no"
}

func formatCLPPlain(amount money.CLP) string {
	return strconv.FormatInt(int64(amount), 10)
}
