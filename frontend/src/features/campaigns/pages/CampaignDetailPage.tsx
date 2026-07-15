import { useRef, useState, type ChangeEvent, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Copy, Pencil } from 'lucide-react'
import {
  useCampaign,
  useCampaignCategories,
  useCampaignGallery,
  useCampaignTransitions,
  useCancelScheduledPublish,
  useCloneCampaign,
  useParticipants,
  usePublishCampaign,
  useUpdateCampaign,
} from '../hooks/useCampaigns'
import { useDebouncedValue } from '../../../shared/lib/useDebouncedValue'
import { errorMessage } from '../../../shared/api/errors'
import { categoryIcon } from '../../../shared/lib/categoryIcons'
import { isValidVideoUrl } from '../../../shared/lib/videoEmbed'
import { Button, Card, Input, ShareButtons, QrCode } from '../../../shared/ui'
import { StatusBadge } from '../components/StatusBadge'
import { typeLabels } from '../typeMeta'
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

// El backend nunca restringió editar título/descripción por estado — solo
// nunca hubo UI para hacerlo, ni siquiera en borrador (el organizador podía
// agregar fotos, pero no corregir un typo en la descripción una vez
// publicada). goal_amount/starts_at/ends_at se reenvían tal cual vienen del
// campaign actual: el endpoint de Update los sobreescribe con lo que llega
// en el body (a diferencia de cover_file_id/publish_at, que sí mantienen su
// valor si viene null) — omitirlos los borraría en cada edición.
function EditCampaignForm({ campaign, onDone }: { campaign: Campaign; onDone: () => void }) {
  const [title, setTitle] = useState(campaign.title)
  const [description, setDescription] = useState(campaign.description)
  const [category, setCategory] = useState(campaign.category)
  const [videoUrl, setVideoUrl] = useState(campaign.video_url ?? '')
  const [error, setError] = useState<string | null>(null)
  const update = useUpdateCampaign(campaign.id)
  const { data: categories } = useCampaignCategories()

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (videoUrl && !isValidVideoUrl(videoUrl)) {
      setError('El link de video debe ser de YouTube o Vimeo.')
      return
    }
    update.mutate(
      {
        title,
        description,
        category,
        goal_amount: campaign.goal_amount,
        starts_at: campaign.starts_at,
        ends_at: campaign.ends_at,
        video_url: videoUrl || undefined,
        clear_video_url: !videoUrl,
      },
      { onSuccess: onDone, onError: (err) => setError(errorMessage(err)) },
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div>
        <label className="mb-1 block text-sm text-text-secondary">Título</label>
        <Input value={title} onChange={(e) => setTitle(e.target.value)} minLength={3} maxLength={120} required />
      </div>
      <div>
        <label className="mb-1 block text-sm text-text-secondary">Descripción</label>
        <textarea
          className="w-full rounded-lg border border-border-default bg-bg-surface px-3 py-2 text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
          rows={5}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
      </div>
      {categories && categories.length > 0 && (
        <div>
          <label className="mb-1 block text-sm text-text-secondary">Categoría</label>
          <div className="flex flex-wrap gap-2">
            {categories.map((c) => {
              const Icon = categoryIcon(c.icon)
              const selected = category === c.key
              return (
                <button
                  key={c.key}
                  type="button"
                  onClick={() => setCategory(c.key)}
                  className={`flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors ${
                    selected
                      ? 'border-brand-hover bg-accent-tint text-brand-hover'
                      : 'border-border-default bg-bg-surface text-text-secondary hover:bg-bg-subtle'
                  }`}
                >
                  <Icon className="h-3.5 w-3.5" strokeWidth={1.75} />
                  {c.label}
                </button>
              )
            })}
          </div>
        </div>
      )}
      <div>
        <label className="mb-1 block text-sm text-text-secondary">
          Video (YouTube o Vimeo, opcional)
        </label>
        <Input
          type="url"
          value={videoUrl}
          onChange={(e) => setVideoUrl(e.target.value)}
          placeholder="https://youtube.com/watch?v=..."
        />
      </div>
      {error && <p className="text-sm text-danger">{error}</p>}
      <div className="flex gap-2">
        <Button type="submit" disabled={update.isPending}>
          {update.isPending ? 'Guardando…' : 'Guardar cambios'}
        </Button>
        <Button type="button" variant="secondary" onClick={onDone} disabled={update.isPending}>
          Cancelar
        </Button>
      </div>
    </form>
  )
}

// RaffleCard: precio/total/vendidos y el campo donde el organizador
// registra el número ganador tras el sorteo externo (Kino/Loto de una
// fecha específica) — Juntalo no sortea nada, solo deja constancia.
function RaffleCard({ campaign }: { campaign: Campaign }) {
  const update = useUpdateCampaign(campaign.id)
  const [winningNumber, setWinningNumber] = useState(
    campaign.raffle_winning_number != null ? String(campaign.raffle_winning_number) : '',
  )
  const [error, setError] = useState<string | null>(null)
  const sold = campaign.raffle_numbers_sold ?? 0
  const total = campaign.raffle_total_numbers ?? 0

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    const n = Number(winningNumber)
    if (!winningNumber || Number.isNaN(n)) {
      setError('Ingresa un número válido.')
      return
    }
    update.mutate(
      {
        title: campaign.title,
        description: campaign.description,
        category: campaign.category,
        goal_amount: campaign.goal_amount,
        starts_at: campaign.starts_at,
        ends_at: campaign.ends_at,
        raffle_winning_number: n,
      },
      { onError: (err) => setError(errorMessage(err)) },
    )
  }

  return (
    <Card className="space-y-3">
      <p className="text-[11px] font-semibold uppercase tracking-wide text-text-secondary">Rifa</p>
      <div className="grid grid-cols-3 gap-2 text-center">
        <div>
          <p className="text-[10px] font-semibold uppercase tracking-wide text-text-secondary">Precio</p>
          <p className="font-display text-sm font-bold tabular-nums">
            ${campaign.raffle_unit_price?.toLocaleString('es-CL')}
          </p>
        </div>
        <div>
          <p className="text-[10px] font-semibold uppercase tracking-wide text-text-secondary">Vendidos</p>
          <p className="font-display text-sm font-bold tabular-nums">{sold}</p>
        </div>
        <div>
          <p className="text-[10px] font-semibold uppercase tracking-wide text-text-secondary">Total</p>
          <p className="font-display text-sm font-bold tabular-nums">{total}</p>
        </div>
      </div>
      <form onSubmit={handleSubmit} className="flex items-end gap-2">
        <div className="flex-1">
          <label className="mb-1 block text-sm text-text-secondary">
            Número ganador (después del sorteo)
          </label>
          <Input
            type="text"
            inputMode="numeric"
            value={winningNumber}
            onChange={(e) => setWinningNumber(e.target.value.replace(/\D/g, ''))}
            placeholder="Ej: 42"
          />
        </div>
        <Button type="submit" disabled={update.isPending}>
          {update.isPending ? 'Guardando…' : 'Guardar'}
        </Button>
      </form>
      {error && <p className="text-sm text-danger">{error}</p>}
    </Card>
  )
}

function CampaignDetailContent({ campaign }: { campaign: Campaign }) {
  const navigate = useNavigate()
  const publish = usePublishCampaign()
  const { pause, resume, finish } = useCampaignTransitions()
  const clone = useCloneCampaign()
  const cancelSchedule = useCancelScheduledPublish()
  const gallery = useCampaignGallery(campaign.id)
  const [isEditing, setIsEditing] = useState(false)
  const [participantSearch, setParticipantSearch] = useState('')
  const [participantStatus, setParticipantStatus] = useState('')
  const debouncedParticipantSearch = useDebouncedValue(participantSearch)
  const {
    participants,
    isLoading: isLoadingParticipants,
    hasNextPage: hasMoreParticipants,
    isFetchingNextPage: isFetchingMoreParticipants,
    fetchNextPage: fetchMoreParticipants,
  } = useParticipants(campaign.id, debouncedParticipantSearch, participantStatus)

  // public_url ya viene absoluta desde el backend (Etapa 4: /c/:slug vive en
  // el dominio del backend, no del frontend — necesario cuando ambos están
  // en dominios distintos, como en este deploy de prueba en Render).
  const publicUrl = campaign.public_url
  const fileInputRef = useRef<HTMLInputElement>(null)

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
          <div className="mt-1 flex items-center gap-2">
            <StatusBadge status={campaign.status} />
            <span className="text-xs text-text-secondary">
              {typeLabels[campaign.type_key] ?? campaign.type_key}
            </span>
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
          <Button
            variant="secondary"
            className="flex items-center gap-1.5"
            disabled={clone.isPending}
            onClick={() =>
              clone.mutate(campaign.id, {
                onSuccess: (cloned) => navigate(`/dashboard/campaigns/${cloned.id}`),
              })
            }
          >
            <Copy className="h-4 w-4" strokeWidth={1.75} />
            {clone.isPending ? 'Clonando…' : 'Clonar'}
          </Button>
        </div>
      </div>

      {campaign.status === 'draft' && campaign.publish_at && (
        <Card className="flex flex-wrap items-center justify-between gap-2 border-brand-hover/30 bg-accent-tint">
          <p className="text-sm text-text-primary">
            Se publicará automáticamente el{' '}
            <span className="font-semibold">
              {new Date(campaign.publish_at).toLocaleString('es-CL', {
                dateStyle: 'medium',
                timeStyle: 'short',
              })}
            </span>
            .
          </p>
          <Button
            variant="secondary"
            disabled={cancelSchedule.isPending}
            onClick={() => cancelSchedule.mutate(campaign.id)}
          >
            {cancelSchedule.isPending ? 'Cancelando…' : 'Cancelar publicación programada'}
          </Button>
        </Card>
      )}

      <TotalsPanel campaign={campaign} />

      {campaign.type_key === 'raffle' && <RaffleCard campaign={campaign} />}

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
          <span className="mb-1 block text-sm text-text-secondary">
            {campaign.images.length > 0 ? 'Agregar más fotos' : 'Agregar fotos (opcional)'}
          </span>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            multiple
            className="hidden"
            onChange={handleAddImages}
          />
          <Button type="button" variant="secondary" onClick={() => fileInputRef.current?.click()}>
            Elegir fotos
          </Button>
        </div>
      </Card>

      <Card className="space-y-3">
        {isEditing ? (
          <EditCampaignForm campaign={campaign} onDone={() => setIsEditing(false)} />
        ) : (
          <div className="flex items-start justify-between gap-2">
            <p className="whitespace-pre-wrap text-text-primary">
              {campaign.description || 'Sin descripción.'}
            </p>
            <button
              type="button"
              onClick={() => setIsEditing(true)}
              className="flex flex-none items-center gap-1 text-xs font-medium text-text-secondary hover:text-text-primary"
            >
              <Pencil className="h-3.5 w-3.5" strokeWidth={1.75} />
              Editar
            </button>
          </div>
        )}
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
          <ParticipantsTable
            items={participants}
            search={participantSearch}
            onSearchChange={setParticipantSearch}
            status={participantStatus}
            onStatusChange={setParticipantStatus}
            hasAnyParticipants={
              participants.length > 0 || !!debouncedParticipantSearch || !!participantStatus
            }
          />
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
