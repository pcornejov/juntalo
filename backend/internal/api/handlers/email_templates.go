package handlers

import "html"

// Plantillas mínimas inline: a este volumen (transaccional, texto corto) no
// vale la pena sumar un motor de templates — HTML plano es más fácil de
// auditar y no depende de nada más.

const resetPasswordSubject = "Recupera tu contraseña de Juntalo"

func resetPasswordHTML(link string) string {
	safeLink := html.EscapeString(link)
	return `<!doctype html>
<html>
<body style="font-family: sans-serif; color: #0f1222; max-width: 480px; margin: 0 auto; padding: 24px;">
  <h1 style="font-size: 20px;">Recupera tu contraseña</h1>
  <p>Recibimos una solicitud para restablecer tu contraseña en Juntalo. Si no fuiste tú, ignora este correo.</p>
  <p style="margin: 24px 0;">
    <a href="` + safeLink + `" style="background: #4338ca; color: #fff; padding: 12px 20px; border-radius: 8px; text-decoration: none; font-weight: 600;">
      Elegir nueva contraseña
    </a>
  </p>
  <p style="color: #5b5f76; font-size: 13px;">Este link expira en 1 hora. Si el botón no funciona, copia este link: ` + safeLink + `</p>
</body>
</html>`
}

const newContributionSubjectPrefix = "Nuevo aporte en "

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
