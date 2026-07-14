import type { ChangeEvent } from 'react'
import { useParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import {
  useCampaign,
  useCampaignGallery,
  useCampaignTransitions,
  useParticipants,
  usePublishCampaign,
} from '../hooks/useCampaigns'
import { Button, Card, ShareButtons, QrCode } from '../../../shared/ui'
import { StatusBadge } from '../components/StatusBadge'
import { TotalsPanel } from '../components/TotalsPanel'
import { ParticipantsTable } from '../components/ParticipantsTable'
import { ExportCsvButton } from '../components/ExportCsvButton'
import type { Campaign } from '../api'

export function CampaignDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { data: campaign, isLoading } = useCampaign(id)

  if (isLoading || !campaign) {
    return <p className="text-text-secondary">Cargando…</p>
  }

  return <CampaignDetailContent campaign={campaign} />
}

function CampaignDetailContent({ campaign }: { campaign: Campaign }) {
  const publish = usePublishCampaign()
  const { pause, resume, finish } = useCampaignTransitions()
  const gallery = useCampaignGallery(campaign.id)
  const {
    participants,
    isLoading: isLoadingParticipants,
    hasNextPage: hasMoreParticipants,
    isFetchingNextPage: isFetchingMoreParticipants,
    fetchNextPage: fetchMoreParticipants,
  } = useParticipants(campaign.id)

  // public_url ya viene absoluta desde el backend (Etapa 4: /c/:slug vive en
  // el dominio del backend, no del frontend — necesario cuando ambos están
  // en dominios distintos, como en este deploy de prueba en Render).
  const publicUrl = campaign.public_url

  function handleAddImages(e: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? [])
    files.forEach((file) => gallery.addImage.mutate(file))
    e.target.value = ''
  }

  function moveImage(index: number, direction: -1 | 1) {
    const target = index + direction
    if (target < 0 || target >= campaign.images.length) return
    const ids = campaign.images.map((img) => img.id)
    ;[ids[index], ids[target]] = [ids[target], ids[index]]
    gallery.reorder.mutate(ids)
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-display text-2xl font-bold tracking-tight">{campaign.title}</h1>
          <div className="mt-1">
            <StatusBadge status={campaign.status} />
          </div>
        </div>
        <div className="flex gap-2">
          {campaign.status === 'draft' && (
            <Button onClick={() => publish.mutate(campaign.id)}>Publicar</Button>
          )}
          {campaign.status === 'active' && (
            <>
              <Button variant="secondary" onClick={() => pause.mutate(campaign.id)}>
                Pausar
              </Button>
              <Button variant="secondary" onClick={() => finish.mutate(campaign.id)}>
                Finalizar
              </Button>
            </>
          )}
          {campaign.status === 'paused' && (
            <>
              <Button onClick={() => resume.mutate(campaign.id)}>Reanudar</Button>
              <Button variant="secondary" onClick={() => finish.mutate(campaign.id)}>
                Finalizar
              </Button>
            </>
          )}
        </div>
      </div>

      <TotalsPanel campaign={campaign} />

      {campaign.status !== 'draft' && (
        <Card className="space-y-3">
          <p className="text-[11px] font-semibold uppercase tracking-wide text-text-secondary">
            Página pública
          </p>
          <ShareButtons url={publicUrl} title={campaign.title} />
          <QrCode url={publicUrl} />
        </Card>
      )}

      <Card className="space-y-3">
        <p className="text-[11px] font-semibold uppercase tracking-wide text-text-secondary">
          Fotos de la campaña
        </p>
        {campaign.images.length > 0 && (
          <div className="grid grid-cols-3 gap-2">
            {campaign.images.map((img, index) => (
              <div key={img.id} className="group relative">
                <img src={img.url} alt="" className="h-24 w-full rounded-lg object-cover" />
                {index === 0 && (
                  <span className="absolute left-1 top-1 rounded-full bg-brand-hover px-1.5 py-0.5 text-[10px] font-semibold text-white">
                    Portada
                  </span>
                )}
                <button
                  type="button"
                  onClick={() => gallery.deleteImage.mutate(img.id)}
                  disabled={gallery.deleteImage.isPending}
                  className="absolute right-1 top-1 flex h-6 w-6 items-center justify-center rounded-full bg-black/60 text-xs text-white opacity-0 transition-opacity group-hover:opacity-100"
                  aria-label="Quitar foto"
                >
                  ✕
                </button>
                <div className="absolute inset-x-1 bottom-1 flex justify-between opacity-0 transition-opacity group-hover:opacity-100">
                  <button
                    type="button"
                    onClick={() => moveImage(index, -1)}
                    disabled={index === 0 || gallery.reorder.isPending}
                    className="flex h-6 w-6 items-center justify-center rounded-full bg-black/60 text-white disabled:opacity-30"
                    aria-label="Mover antes"
                  >
                    <ChevronLeft className="h-3.5 w-3.5" strokeWidth={2} />
                  </button>
                  <button
                    type="button"
                    onClick={() => moveImage(index, 1)}
                    disabled={index === campaign.images.length - 1 || gallery.reorder.isPending}
                    className="flex h-6 w-6 items-center justify-center rounded-full bg-black/60 text-white disabled:opacity-30"
                    aria-label="Mover después"
                  >
                    <ChevronRight className="h-3.5 w-3.5" strokeWidth={2} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
        <div>
          <label className="mb-1 block text-sm text-text-secondary">
            {campaign.images.length > 0 ? 'Agregar más fotos' : 'Agregar fotos (opcional)'}
          </label>
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            multiple
            onChange={handleAddImages}
          />
        </div>
      </Card>

      <Card>
        <p className="whitespace-pre-wrap text-text-primary">
          {campaign.description || 'Sin descripción.'}
        </p>
      </Card>

      <Card className="space-y-4">
        <div className="flex items-center justify-between">
          <p className="text-[11px] font-semibold uppercase tracking-wide text-text-secondary">
            Participantes
          </p>
          <ExportCsvButton campaignId={campaign.id} />
        </div>
        {isLoadingParticipants ? (
          <p className="text-sm text-text-secondary">Cargando…</p>
        ) : (
          <ParticipantsTable items={participants} />
        )}
        {hasMoreParticipants && (
          <div className="text-center">
            <Button
              variant="secondary"
              onClick={() => fetchMoreParticipants()}
              disabled={isFetchingMoreParticipants}
            >
              {isFetchingMoreParticipants ? 'Cargando…' : 'Cargar más'}
            </Button>
          </div>
        )}
      </Card>
    </div>
  )
}
