import type { ChangeEvent } from 'react'
import { useParams } from 'react-router-dom'
import {
  useCampaign,
  useCampaignTransitions,
  useParticipants,
  usePublishCampaign,
  useUploadCover,
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
  const uploadCover = useUploadCover(campaign)
  const { data: participants, isLoading: isLoadingParticipants } = useParticipants(campaign.id)

  // public_url ya viene absoluta desde el backend (Etapa 4: /c/:slug vive en
  // el dominio del backend, no del frontend — necesario cuando ambos están
  // en dominios distintos, como en este deploy de prueba en Render).
  const publicUrl = campaign.public_url

  function handleCoverChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) uploadCover.mutate(file)
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{campaign.title}</h1>
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
          <p className="text-sm text-text-secondary">Página pública</p>
          <ShareButtons url={publicUrl} title={campaign.title} />
          <QrCode url={publicUrl} />
        </Card>
      )}

      <Card className="space-y-3">
        {campaign.cover_url && (
          <img src={campaign.cover_url} alt="" className="h-40 w-full rounded-lg object-cover" />
        )}
        <div>
          <label className="mb-1 block text-sm text-text-secondary">
            {campaign.cover_url ? 'Cambiar imagen' : 'Agregar imagen (opcional)'}
          </label>
          <input type="file" accept="image/jpeg,image/png,image/webp" onChange={handleCoverChange} />
        </div>
      </Card>

      <Card>
        <p className="whitespace-pre-wrap text-text-primary">
          {campaign.description || 'Sin descripción.'}
        </p>
      </Card>

      <Card className="space-y-4">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium text-text-secondary">Participantes</p>
          <ExportCsvButton campaignId={campaign.id} />
        </div>
        {isLoadingParticipants ? (
          <p className="text-sm text-text-secondary">Cargando…</p>
        ) : (
          <ParticipantsTable items={participants ?? []} />
        )}
      </Card>
    </div>
  )
}
