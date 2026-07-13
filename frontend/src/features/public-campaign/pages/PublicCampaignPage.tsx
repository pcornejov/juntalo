import { useParams } from 'react-router-dom'
import { usePublicCampaign } from '../hooks/usePublicCampaign'
import { formatCLP } from '../../../shared/lib/clp'
import { absoluteUrl } from '../../../shared/lib/share'
import { Button, Card, Progress, ShareButtons, QrCode } from '../../../shared/ui'

// Diseñada mobile-first a 375px: es la página que se abre desde el in-app
// browser de WhatsApp (Etapa 2 §3, Etapa 5 §3).
export function PublicCampaignPage() {
  const { slug } = useParams<{ slug: string }>()
  const { data: campaign, isLoading, isError } = usePublicCampaign(slug)

  if (isLoading) {
    return <div className="p-6 text-center text-text-secondary">Cargando…</div>
  }

  if (isError || !campaign) {
    return (
      <div className="p-6 text-center text-text-secondary">
        Esta campaña no existe o ya no está disponible.
      </div>
    )
  }

  const publicUrl = absoluteUrl(`/public/${slug}`)

  return (
    <div className="mx-auto max-w-md pb-24">
      {campaign.cover_url ? (
        <img src={campaign.cover_url} alt={campaign.title} className="h-56 w-full object-cover" />
      ) : (
        <div className="h-40 w-full bg-bg-subtle" />
      )}

      <div className="space-y-4 p-4">
        <h1 className="text-xl font-semibold">{campaign.title}</h1>

        {campaign.goal_amount ? (
          <>
            <Progress value={campaign.totals.raised_gross} max={campaign.goal_amount} />
            <p className="text-sm text-text-secondary">
              {formatCLP(campaign.totals.raised_gross)} de {formatCLP(campaign.goal_amount)}
            </p>
          </>
        ) : (
          <p className="text-sm text-text-secondary">
            {formatCLP(campaign.totals.raised_gross)} recaudados
          </p>
        )}

        <p className="text-sm text-text-secondary">
          {campaign.totals.contributor_count} {campaign.unit}
        </p>

        <p className="whitespace-pre-wrap text-text-primary">{campaign.description}</p>

        <Card className="space-y-3 text-center">
          <p className="text-sm text-text-secondary">Comparte esta campaña</p>
          <ShareButtons url={publicUrl} title={campaign.title} />
          <QrCode url={publicUrl} />
        </Card>
      </div>

      <div className="fixed inset-x-0 bottom-0 border-t border-border-default bg-bg-surface p-4">
        <Button className="w-full" disabled title="Los pagos se habilitan en el próximo hito">
          {campaign.cta}
        </Button>
      </div>
    </div>
  )
}
