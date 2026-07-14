import { useState, type ChangeEvent, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  useCampaignCategories,
  useCampaignTypes,
  useCreateCampaign,
  usePublishCampaign,
} from '../hooks/useCampaigns'
import { campaignArchetypes } from '../archetypes'
import { typeIcons } from '../typeMeta'
import { errorMessage } from '../../../shared/api/errors'
import { formatCLP } from '../../../shared/lib/clp'
import { categoryIcon } from '../../../shared/lib/categoryIcons'
import { isValidVideoUrl } from '../../../shared/lib/videoEmbed'
import { Button, Card, Input } from '../../../shared/ui'

// El formulario crea Y publica en un solo paso (Etapa 1: crear y compartir en <2 min).
// La imagen queda fuera a propósito: es opcional y se agrega después desde el detalle.
export function CampaignFormPage() {
  const navigate = useNavigate()
  const { data: types, isLoading: loadingTypes } = useCampaignTypes()
  const { data: categories } = useCampaignCategories()
  const createCampaign = useCreateCampaign()
  const publishCampaign = usePublishCampaign()

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [goalAmount, setGoalAmount] = useState('')
  const [raffleUnitPrice, setRaffleUnitPrice] = useState('')
  const [raffleTotalNumbers, setRaffleTotalNumbers] = useState('')
  const [category, setCategory] = useState<string>('')
  const [videoUrl, setVideoUrl] = useState('')
  const [selectedArchetype, setSelectedArchetype] = useState<string | null>(null)
  const [typeKey, setTypeKey] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [publishMode, setPublishMode] = useState<'now' | 'later'>('now')
  const [publishAt, setPublishAt] = useState('')

  // El primer tipo de la lista queda como default hasta que el organizador
  // elija otro — la lista viene en orden estable desde el backend.
  const selectedType = types?.find((t) => t.key === typeKey) ?? types?.[0]
  const defaultType = selectedType

  const isRaffle = selectedType?.key === 'raffle'

  function handleGoalAmountChange(e: ChangeEvent<HTMLInputElement>) {
    setGoalAmount(e.target.value.replace(/\D/g, ''))
  }

  function handleRaffleUnitPriceChange(e: ChangeEvent<HTMLInputElement>) {
    setRaffleUnitPrice(e.target.value.replace(/\D/g, ''))
  }

  function handleRaffleTotalNumbersChange(e: ChangeEvent<HTMLInputElement>) {
    setRaffleTotalNumbers(e.target.value.replace(/\D/g, ''))
  }

  function applyArchetype(id: string) {
    const archetype = campaignArchetypes.find((a) => a.id === id)
    if (!archetype) return
    setSelectedArchetype(id)
    setTitle(archetype.title)
    setDescription(archetype.description)
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!defaultType) return
    setError(null)
    if (publishMode === 'later' && !publishAt) {
      setError('Elige una fecha de publicación.')
      return
    }
    if (videoUrl && !isValidVideoUrl(videoUrl)) {
      setError('El link de video debe ser de YouTube o Vimeo.')
      return
    }
    if (isRaffle && (!raffleUnitPrice || !raffleTotalNumbers)) {
      setError('Ingresa el precio por número y el total de números de la rifa.')
      return
    }
    setIsSubmitting(true)
    try {
      const created = await createCampaign.mutateAsync({
        type_key: defaultType.key,
        category: category || undefined,
        title,
        description: description || undefined,
        goal_amount: goalAmount ? Number(goalAmount) : undefined,
        publish_at: publishMode === 'later' ? new Date(publishAt).toISOString() : undefined,
        video_url: videoUrl || undefined,
        raffle_unit_price: isRaffle ? Number(raffleUnitPrice) : undefined,
        raffle_total_numbers: isRaffle ? Number(raffleTotalNumbers) : undefined,
      })
      if (publishMode === 'now') {
        await publishCampaign.mutateAsync(created.id)
      }
      navigate(`/dashboard/campaigns/${created.id}`)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  if (loadingTypes) {
    return <p className="text-text-secondary">Cargando…</p>
  }

  return (
    <div className="mx-auto max-w-lg">
      <h1 className="mb-6 font-display text-2xl font-bold tracking-tight">Nueva campaña</h1>

      {types && types.length > 1 && (
        <div className="mb-6">
          <p className="mb-2 text-sm text-text-secondary">Tipo de campaña</p>
          <div className="flex flex-wrap gap-2">
            {types.map((t) => {
              const Icon = typeIcons[t.key]
              const selected = selectedType?.key === t.key
              return (
                <button
                  key={t.key}
                  type="button"
                  onClick={() => setTypeKey(t.key)}
                  className={`flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors ${
                    selected
                      ? 'border-brand-hover bg-accent-tint text-brand-hover'
                      : 'border-border-default bg-bg-surface text-text-secondary hover:bg-bg-subtle'
                  }`}
                >
                  {Icon && <Icon className="h-3.5 w-3.5" strokeWidth={1.75} />}
                  {t.name}
                </button>
              )
            })}
          </div>
        </div>
      )}

      {selectedType?.key === 'collection' && (
        <div className="mb-6">
          <p className="mb-2 text-sm text-text-secondary">
            Empieza con una plantilla (opcional) — igual puedes editar todo después
          </p>
          <div className="grid grid-cols-3 gap-2 sm:grid-cols-4">
            {campaignArchetypes.map((a) => (
              <button
                key={a.id}
                type="button"
                onClick={() => applyArchetype(a.id)}
                className={`flex flex-col items-center gap-1.5 rounded-xl border p-3 text-center transition-colors ${
                  selectedArchetype === a.id
                    ? 'border-brand-hover bg-accent-tint'
                    : 'border-border-default bg-bg-surface hover:bg-bg-subtle'
                }`}
              >
                <a.icon
                  className={`h-5 w-5 ${selectedArchetype === a.id ? 'text-brand-hover' : 'text-text-secondary'}`}
                  strokeWidth={1.75}
                />
                <span className="text-[11px] font-medium leading-tight text-text-primary">{a.label}</span>
              </button>
            ))}
          </div>
        </div>
      )}

      {categories && categories.length > 0 && (
        <div className="mb-6">
          <p className="mb-2 text-sm text-text-secondary">Categoría (opcional)</p>
          <div className="flex flex-wrap gap-2">
            {categories.map((c) => {
              const Icon = categoryIcon(c.icon)
              const selected = category === c.key
              return (
                <button
                  key={c.key}
                  type="button"
                  onClick={() => setCategory(selected ? '' : c.key)}
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

      <Card>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm text-text-secondary">Título</label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Ej: Viaje de estudios cuarto medio"
              required
              minLength={3}
              maxLength={120}
            />
          </div>
          <div>
            <label className="mb-1 block text-sm text-text-secondary">Descripción</label>
            <textarea
              className="w-full rounded-lg border border-border-default bg-bg-surface px-3 py-2 text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
              rows={4}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Cuenta de qué se trata tu campaña"
            />
          </div>
          {isRaffle ? (
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-sm text-text-secondary">
                  Precio por número (CLP)
                </label>
                <Input
                  type="text"
                  inputMode="numeric"
                  value={raffleUnitPrice ? formatCLP(Number(raffleUnitPrice)) : ''}
                  onChange={handleRaffleUnitPriceChange}
                  placeholder="$1.000"
                  required
                />
              </div>
              <div>
                <label className="mb-1 block text-sm text-text-secondary">Total de números</label>
                <Input
                  type="text"
                  inputMode="numeric"
                  value={raffleTotalNumbers}
                  onChange={handleRaffleTotalNumbersChange}
                  placeholder="100"
                  required
                />
              </div>
            </div>
          ) : (
            <div>
              <label className="mb-1 block text-sm text-text-secondary">
                Meta (CLP, opcional)
              </label>
              <Input
                type="text"
                inputMode="numeric"
                value={goalAmount ? formatCLP(Number(goalAmount)) : ''}
                onChange={handleGoalAmountChange}
                placeholder="$500.000"
              />
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
          <div>
            <label className="mb-2 block text-sm text-text-secondary">Publicación</label>
            <div className="flex gap-4 text-sm">
              <label className="flex items-center gap-1.5">
                <input
                  type="radio"
                  name="publishMode"
                  checked={publishMode === 'now'}
                  onChange={() => setPublishMode('now')}
                />
                Publicar ahora
              </label>
              <label className="flex items-center gap-1.5">
                <input
                  type="radio"
                  name="publishMode"
                  checked={publishMode === 'later'}
                  onChange={() => setPublishMode('later')}
                />
                Programar para más adelante
              </label>
            </div>
            {publishMode === 'later' && (
              <Input
                type="datetime-local"
                value={publishAt}
                onChange={(e) => setPublishAt(e.target.value)}
                min={new Date().toISOString().slice(0, 16)}
                className="mt-2"
                required
              />
            )}
          </div>
          {error && <p className="text-sm text-danger">{error}</p>}
          <Button type="submit" disabled={isSubmitting || !defaultType} className="w-full">
            {isSubmitting
              ? 'Creando…'
              : publishMode === 'later'
                ? 'Crear y programar publicación'
                : 'Crear y publicar'}
          </Button>
        </form>
      </Card>
    </div>
  )
}
