import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search, Users } from 'lucide-react'
import { useExploreCampaigns } from '../hooks/usePublicCampaign'
import { useDebouncedValue } from '../../../shared/lib/useDebouncedValue'
import { formatCLP } from '../../../shared/lib/clp'
import { Button, Footer, Input, Progress, ThemeToggle } from '../../../shared/ui'
import type { PublicCampaign } from '../api'

function CampaignExploreCard({ campaign }: { campaign: PublicCampaign }) {
  return (
    <Link to={`/public/${campaign.slug}`}>
      <div className="h-full overflow-hidden rounded-xl border border-border-default bg-bg-surface shadow-sm transition-shadow hover:shadow-md">
        <div className="h-32 w-full bg-bg-subtle">
          {campaign.cover_url && (
            <img src={campaign.cover_url} alt="" className="h-32 w-full object-cover" />
          )}
        </div>
        <div className="space-y-3 p-4">
          <h3 className="font-display text-sm font-bold leading-tight">{campaign.title}</h3>
          <p className="text-sm text-text-secondary">
            <span className="font-semibold text-brand-hover">
              {formatCLP(campaign.totals.raised_gross)}
            </span>
            {campaign.goal_amount ? ` de ${formatCLP(campaign.goal_amount)}` : ' recaudados'}
          </p>
          {campaign.goal_amount && (
            <Progress value={campaign.totals.raised_gross} max={campaign.goal_amount} />
          )}
          <div className="flex items-center justify-between">
            <p className="flex items-center gap-1 text-xs text-text-secondary">
              <Users className="h-3 w-3" strokeWidth={1.75} />
              {campaign.totals.contributor_count} {campaign.unit}
            </p>
            <span className="text-xs font-semibold text-brand-hover">{campaign.cta} →</span>
          </div>
        </div>
      </div>
    </Link>
  )
}

// Sección pública de descubrimiento: antes solo se podía llegar a una
// campaña con el link exacto que compartió el organizador — no había forma
// de simplemente "ver qué campañas hay" para aportar en alguna.
export function ExploreCampaignsPage() {
  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const { campaigns, isLoading, hasNextPage, isFetchingNextPage, fetchNextPage } =
    useExploreCampaigns(debouncedSearch)

  return (
    <div className="min-h-screen bg-bg-subtle">
      <header className="mx-auto flex max-w-5xl items-center justify-between px-6 py-5">
        <Link to="/" className="font-display text-lg font-bold tracking-tight">
          Juntalo
        </Link>
        <div className="flex items-center gap-1">
          <ThemeToggle />
          <Link to="/login" className="text-sm font-medium text-text-secondary hover:text-text-primary">
            Ingresar
          </Link>
        </div>
      </header>

      <div className="mx-auto max-w-5xl px-6 pb-16">
        <h1 className="mb-2 font-display text-2xl font-bold tracking-tight">Campañas activas</h1>
        <p className="mb-6 text-text-secondary">
          Explora campañas que ya están recibiendo aportes y súmate a la que quieras.
        </p>

        <div className="relative mb-6 max-w-md">
          <Search
            className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-secondary"
            strokeWidth={1.75}
          />
          <Input
            type="search"
            placeholder="Buscar por título…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        {isLoading && <p className="text-text-secondary">Cargando…</p>}

        {!isLoading && campaigns.length === 0 && (
          <p className="text-text-secondary">
            {debouncedSearch
              ? 'No hay campañas activas que coincidan con la búsqueda.'
              : 'Todavía no hay campañas activas. Vuelve pronto.'}
          </p>
        )}

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {campaigns.map((c) => (
            <CampaignExploreCard key={c.slug} campaign={c} />
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

      <Footer />
    </div>
  )
}
