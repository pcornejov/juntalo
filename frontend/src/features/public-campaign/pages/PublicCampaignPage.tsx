import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { Calendar, MapPin, Share2, Link2, ArrowRight, ShieldCheck, Users } from 'lucide-react'
import { usePublicCampaign } from '../hooks/usePublicCampaign'
import { useContributionStatus } from '../hooks/useContribute'
import { formatCLP } from '../../../shared/lib/clp'
import { relativeDate } from '../../../shared/lib/date'
import { Button, Card, Carousel, Progress, ShareButtons, QrCode, Footer } from '../../../shared/ui'
import { ContributeSheet } from '../components/ContributeSheet'
import { OrganizerCard } from '../components/OrganizerCard'
import { ContributeSuccessPanel } from './ContributeSuccessPage'
import type { StartContributionResult } from '../api'

// Diseñada mobile-first a 375px: es la página que se abre desde el in-app
// browser de WhatsApp (Etapa 2 §3, Etapa 5 §3).
//
// Dirección visual: B1 "Índigo eléctrico" — un solo acento en dos tonos
// (--color-brand para progreso/cifras/badges, --color-brand-hover para el
// CTA principal), Sora en titulares y cifras grandes, Inter en el resto.
export function PublicCampaignPage() {
  const { slug } = useParams<{ slug: string }>()
  const { data: campaign, isLoading, isError } = usePublicCampaign(slug)
  const [isSheetOpen, setIsSheetOpen] = useState(false)
  const [result, setResult] = useState<StartContributionResult | null>(null)
  const { data: statusData } = useContributionStatus(
    slug ?? '',
    result?.contribution_id,
    result?.payment.status,
  )
  const contributionStatus = statusData?.status ?? result?.payment.status
  // Mientras el pago del aporte recién hecho sigue 'pending', el total y el
  // contador de aportantes que devuelve el backend todavía no lo incluyen
  // (solo cuentan pagos confirmados) — se marcan como "actualizando" en vez
  // de mostrar una cifra que el propio usuario sabe que está desactualizada.
  const isConfirmingContribution = Boolean(result) && contributionStatus === 'pending'

  if (isLoading) {
    return <div className="p-6 text-center text-text-secondary">Cargando…</div>
  }

  if (isError || !campaign || !slug) {
    return (
      <div className="p-6 text-center text-text-secondary">
        Esta campaña no existe o ya no está disponible.
      </div>
    )
  }

  // public_url viene del backend y apunta a /c/:slug (con OG tags para
  // WhatsApp/Facebook/etc.) — no se reconstruye en el frontend para que
  // re-compartir desde aquí siga mostrando preview con imagen y título.
  const publicUrl = campaign.public_url
  const canContribute = campaign.status === 'active'
  const hasGoal = Boolean(campaign.goal_amount)
  // Campos aditivos que el backend actual no siempre envía: la UI se
  // degrada mostrando menos, nunca inventando un dato que no llegó.
  const hasFacts = Boolean(campaign.created_at) || Boolean(campaign.location)

  return (
    <div className="mx-auto max-w-md pb-28">
      <div className="relative">
        {campaign.images.length > 0 ? (
          <Carousel images={campaign.images} alt={campaign.title} className="h-56 w-full object-cover" />
        ) : (
          <div className="h-40 w-full bg-bg-subtle" />
        )}
        {campaign.images.length > 0 && (
          <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/80 via-black/0 to-black/0" />
        )}
        {campaign.is_verified && (
          <div className="absolute bottom-3 left-4 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-white">
            <ShieldCheck className="h-3.5 w-3.5" strokeWidth={1.75} />
            Campaña verificada
          </div>
        )}
      </div>

      <div className="space-y-5 p-4">
        <div className="space-y-3">
          <h1 className="text-balance font-display text-2xl font-bold leading-tight tracking-tight">
            {campaign.title}
          </h1>

          {hasFacts && (
            <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs font-medium text-text-secondary">
              {campaign.created_at && (
                <span className="flex items-center gap-1.5">
                  <Calendar className="h-3.5 w-3.5" strokeWidth={1.75} />
                  {relativeDate(campaign.created_at)}
                </span>
              )}
              {campaign.location && (
                <span className="flex items-center gap-1.5">
                  <MapPin className="h-3.5 w-3.5" strokeWidth={1.75} />
                  {campaign.location}
                </span>
              )}
            </div>
          )}
        </div>

        <div>
          {isConfirmingContribution && (
            <p className="mb-1.5 flex items-center gap-1.5 text-[11px] font-medium text-text-secondary">
              <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-brand" />
              Actualizando con tu aporte…
            </p>
          )}
          <div className={isConfirmingContribution ? 'animate-pulse opacity-50' : undefined}>
            {hasGoal ? (
              <div className="overflow-hidden rounded-2xl border border-border-default">
                <div className="grid grid-cols-3 divide-x divide-border-default">
                  <div className="p-3">
                    <p className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-text-secondary">
                      Recaudado
                    </p>
                    <p className="font-display text-base font-bold tabular-nums text-brand-hover">
                      {formatCLP(campaign.totals.raised_gross)}
                    </p>
                  </div>
                  <div className="p-3">
                    <p className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-text-secondary">
                      Meta
                    </p>
                    <p className="font-display text-base font-bold tabular-nums">
                      {formatCLP(campaign.goal_amount!)}
                    </p>
                  </div>
                  <div className="p-3">
                    <p className="mb-1 flex items-center gap-1 text-[10px] font-semibold uppercase tracking-wide text-text-secondary">
                      <Users className="h-2.5 w-2.5" strokeWidth={1.75} />
                      {campaign.unit}
                    </p>
                    <p className="font-display text-base font-bold tabular-nums">
                      {campaign.totals.contributor_count}
                    </p>
                  </div>
                </div>
                <div className="px-3 pb-3 pt-1">
                  <Progress value={campaign.totals.raised_gross} max={campaign.goal_amount!} />
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-between rounded-2xl border border-border-default p-3">
                <div>
                  <p className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-text-secondary">
                    Recaudado
                  </p>
                  <p className="font-display text-lg font-bold tabular-nums text-brand-hover">
                    {formatCLP(campaign.totals.raised_gross)}
                  </p>
                </div>
                <div className="text-right">
                  <p className="mb-1 flex items-center justify-end gap-1 text-[10px] font-semibold uppercase tracking-wide text-text-secondary">
                    <Users className="h-2.5 w-2.5" strokeWidth={1.75} />
                    {campaign.unit}
                  </p>
                  <p className="font-display text-lg font-bold tabular-nums">
                    {campaign.totals.contributor_count}
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>

        <div>
          <p className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-text-secondary">
            Sobre esta campaña
          </p>
          <p className="whitespace-pre-wrap text-[15px] leading-relaxed text-text-primary">
            {campaign.description}
          </p>
        </div>

        {campaign.organizer_name && (
          <OrganizerCard
            name={campaign.organizer_name}
            campaignCount={campaign.organizer_campaign_count}
            isVerified={campaign.is_verified}
          />
        )}

        <hr className="border-border-default" />

        <Card className="space-y-3">
          <p className="flex items-center gap-2 font-display text-sm font-bold">
            <Share2 className="h-4 w-4 text-text-secondary" strokeWidth={1.75} />
            Compartir campaña
          </p>
          <ShareButtons url={publicUrl} title={campaign.title} />
          <div className="flex items-center gap-2 rounded-lg border border-dashed border-border-default px-3 py-2">
            <Link2 className="h-3.5 w-3.5 flex-none text-text-secondary" strokeWidth={1.75} />
            <span className="flex-1 truncate text-xs text-text-secondary">{publicUrl}</span>
          </div>
          <QrCode url={publicUrl} />
        </Card>
      </div>

      <Footer />

      <div className="fixed inset-x-0 bottom-0 border-t border-border-default bg-bg-surface p-4">
        <Button
          variant="cta"
          className="flex w-full items-center justify-center gap-2"
          disabled={!canContribute}
          onClick={() => setIsSheetOpen(true)}
        >
          {canContribute ? campaign.cta : 'Campaña no disponible'}
          {canContribute && <ArrowRight className="h-4 w-4" strokeWidth={1.75} />}
        </Button>
      </div>

      {isSheetOpen && (
        <ContributeSheet
          slug={slug}
          cta={campaign.cta}
          onClose={() => setIsSheetOpen(false)}
          onSuccess={(r) => {
            setIsSheetOpen(false)
            setResult(r)
          }}
        />
      )}

      {result && (
        <ContributeSuccessPanel
          status={contributionStatus ?? result.payment.status}
          onClose={() => setResult(null)}
        />
      )}
    </div>
  )
}
