import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { useCampaigns } from '../hooks/useCampaigns'
import { Button } from '../../../shared/ui'
import { CampaignCard } from '../components/CampaignCard'

export function CampaignListPage() {
  const { campaigns, isLoading, hasNextPage, isFetchingNextPage, fetchNextPage } = useCampaigns()

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="font-display text-2xl font-bold tracking-tight">Mis campañas</h1>
        <Link to="/dashboard/campaigns/new">
          <Button className="flex items-center gap-1.5">
            <Plus className="h-4 w-4" strokeWidth={2} />
            Nueva campaña
          </Button>
        </Link>
      </div>

      {isLoading && <p className="text-text-secondary">Cargando…</p>}

      {!isLoading && campaigns.length === 0 && (
        <p className="text-text-secondary">
          Aún no tienes campañas.{' '}
          <Link to="/dashboard/campaigns/new" className="text-brand-hover hover:underline">
            Crea la primera
          </Link>
          .
        </p>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {campaigns.map((c) => (
          <CampaignCard key={c.id} campaign={c} />
        ))}
      </div>

      {hasNextPage && (
        <div className="mt-6 text-center">
          <Button variant="secondary" onClick={() => fetchNextPage()} disabled={isFetchingNextPage}>
            {isFetchingNextPage ? 'Cargando…' : 'Cargar más'}
          </Button>
        </div>
      )}
    </div>
  )
}
