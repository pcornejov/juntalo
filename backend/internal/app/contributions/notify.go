package contributions

import (
	"context"
	"html"
	"log"
	"strconv"
	"strings"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
)

// notifyOrganizer manda un email best-effort al dueño de la campaña cuando
// un pago se confirma. Se llama tanto desde StartService (modo mock
// "instant", que confirma en el mismo request) como desde ConfirmService
// (webhook/"deferred") — son los dos únicos lugares donde un pago puede
// pasar a confirmed (Etapa 4 §5). Un error de envío no debe tumbar el flujo
// de pago, así que solo se loguea.
func notifyOrganizer(
	ctx context.Context,
	contributions app.ContributionRepository,
	campaigns app.CampaignRepository,
	contributors app.ContributorRepository,
	orgs app.OrganizationRepository,
	emailSender app.EmailSender,
	frontendURL string,
	pay payment.Payment,
) {
	contrib, found, err := contributions.GetByID(ctx, pay.ContributionID)
	if err != nil || !found {
		return
	}
	camp, found, err := campaigns.GetByID(ctx, contrib.CampaignID)
	if err != nil || !found {
		return
	}
	ownerEmail, _, err := orgs.GetOwnerEmail(ctx, camp.OrganizationID)
	if err != nil || ownerEmail == "" {
		return
	}

	// El organizador siempre ve el nombre real, aunque el aporte sea
	// anónimo de cara al público (Etapa 1 riesgo 15: el dashboard siempre
	// muestra los datos completos).
	contributorName := "Alguien"
	if c, found, err := contributors.GetByID(ctx, contrib.ContributorID); err == nil && found && c.FullName != "" {
		contributorName = c.FullName
	}

	link := frontendURL + "/dashboard/campaigns/" + camp.ID.String()
	subject := "Nuevo aporte en " + camp.Title
	body := newContributionHTML(camp.Title, contributorName, formatCLP(pay.AmountGross), link)

	if err := emailSender.Send(ctx, ownerEmail, subject, body); err != nil {
		log.Printf("notify organizer: email send failed: %v", err)
	}
}

func formatCLP(amount money.CLP) string {
	s := strconv.FormatInt(int64(amount), 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, c)
	}
	result := "$" + string(out)
	if neg {
		result = "-" + result
	}
	return result
}

func newContributionHTML(campaignTitle, contributorName, amountFormatted, dashboardLink string) string {
	title := html.EscapeString(campaignTitle)
	name := html.EscapeString(contributorName)
	amount := html.EscapeString(amountFormatted)
	link := html.EscapeString(dashboardLink)
	return `<!doctype html>
<html>
<body style="font-family: sans-serif; color: #0f1222; max-width: 480px; margin: 0 auto; padding: 24px;">
  <h1 style="font-size: 20px;">¡Nuevo aporte confirmado!</h1>
  <p><strong>` + name + `</strong> aportó <strong>` + amount + `</strong> a tu campaña "` + title + `".</p>
  <p style="margin: 24px 0;">
    <a href="` + link + `" style="background: #4338ca; color: #fff; padding: 12px 20px; border-radius: 8px; text-decoration: none; font-weight: 600;">
      Ver campaña
    </a>
  </p>
</body>
</html>`
}
