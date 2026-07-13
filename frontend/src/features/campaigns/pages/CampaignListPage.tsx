import { Link } from 'react-router-dom'
import { useCampaigns } from '../hooks/useCampaigns'
import { Button } from '../../../shared/ui'
import { CampaignCard } from '../components/CampaignCard'

export function CampaignListPage() {
  const { data: campaigns, isLoading } = useCampaigns()

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Mis campañas</h1>
        <Link to="/dashboard/campaigns/new">
          <Button>Nueva campaña</Button>
        </Link>
      </div>

      {isLoading && <p className="text-text-secondary">Cargando…</p>}

      {campaigns && campaigns.length === 0 && (
        <p className="text-text-secondary">
          Aún no tienes campañas.{' '}
          <Link to="/dashboard/campaigns/new" className="text-brand hover:underline">
            Crea la primera
          </Link>
          .
        </p>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {campaigns?.map((c) => (
          <CampaignCard key={c.id} campaign={c} />
        ))}
      </div>
    </div>
  )
}
