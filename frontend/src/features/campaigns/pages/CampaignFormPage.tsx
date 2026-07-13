import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCampaignTypes, useCreateCampaign, usePublishCampaign } from '../hooks/useCampaigns'
import { errorMessage } from '../../../shared/api/errors'
import { Button, Card, Input } from '../../../shared/ui'

// El formulario crea Y publica en un solo paso (Etapa 1: crear y compartir en <2 min).
// La imagen queda fuera a propósito: es opcional y se agrega después desde el detalle.
export function CampaignFormPage() {
  const navigate = useNavigate()
  const { data: types, isLoading: loadingTypes } = useCampaignTypes()
  const createCampaign = useCreateCampaign()
  const publishCampaign = usePublishCampaign()

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [goalAmount, setGoalAmount] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const defaultType = types?.[0]

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!defaultType) return
    setError(null)
    setIsSubmitting(true)
    try {
      const created = await createCampaign.mutateAsync({
        type_key: defaultType.key,
        title,
        description: description || undefined,
        goal_amount: goalAmount ? Number(goalAmount) : undefined,
      })
      await publishCampaign.mutateAsync(created.id)
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
      <h1 className="mb-6 text-2xl font-semibold">Nueva campaña</h1>
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
          <div>
            <label className="mb-1 block text-sm text-text-secondary">
              Meta (CLP, opcional)
            </label>
            <Input
              type="number"
              min={0}
              value={goalAmount}
              onChange={(e) => setGoalAmount(e.target.value)}
              placeholder="500000"
            />
          </div>
          {error && <p className="text-sm text-danger">{error}</p>}
          <Button type="submit" disabled={isSubmitting || !defaultType} className="w-full">
            {isSubmitting ? 'Creando…' : 'Crear y publicar'}
          </Button>
        </form>
      </Card>
    </div>
  )
}
