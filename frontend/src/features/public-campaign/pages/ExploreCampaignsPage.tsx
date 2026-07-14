import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search, Sparkles, Users } from 'lucide-react'
import { useExploreCampaigns, useFeaturedCampaign } from '../hooks/usePublicCampaign'
import { useCampaignCategories } from '../../campaigns/hooks/useCampaigns'
import { useDebouncedValue } from '../../../shared/lib/useDebouncedValue'
import { formatCLP } from '../../../shared/lib/clp'
import { categoryIcon } from '../../../shared/lib/categoryIcons'
import { Button, Footer, Input, Progress, ThemeToggle } from '../../../shared/ui'
import { CampaignExploreCard, VerifiedBadge } from '../components/CampaignExploreCard'
import type { PublicCampaign } from '../api'

// Tarjeta destacada estilo "más caliente" (inspirado en Vaki): la campaña
// activa con el aporte confirmado más reciente de toda la plataforma, no un
// campo que el organizador marca manualmente.
function FeaturedCampaignCard({ campaign }: { campaign: PublicCampaign }) {
  return (
    <Link to={`/public/${campaign.slug}`} className="mb-6 block">
      <div className="overflow-hidden rounded-2xl border border-border-default bg-bg-surface shadow-sm transition-shadow hover:shadow-md sm:flex">
        <div className="h-40 w-full bg-bg-subtle sm:h-auto sm:w-64 sm:flex-none">
          {campaign.cover_url && (
            <img src={campaign.cover_url} alt="" className="h-full w-full object-cover" />
          )}
        </div>
        <div className="flex-1 space-y-3 p-5">
          <span className="flex w-fit items-center gap-1.5 rounded-full bg-accent-tint px-2.5 py-1 text-[11px] font-bold uppercase tracking-wide text-brand-hover">
            <Sparkles className="h-3 w-3" strokeWidth={1.75} />
            Destacada
          </span>
          <div className="flex items-start justify-between gap-2">
            <h3 className="font-display text-lg font-bold leading-tight">{campaign.title}</h3>
            {campaign.is_verified && <VerifiedBadge />}
          </div>
          <p className="text-sm text-text-secondary">
            <span className="font-semibold text-brand-hover">
              {formatCLP(campaign.totals.raised_gross)}
            </span>
            {campaign.goal_amount ? ` de ${formatCLP(campaign.goal_amount)}` : ' recaudados'}
          </p>
          {campaign.goal_amount && (
            <Progress value={campaign.totals.raised_gross} max={campaign.goal_amount} />
          )}
          <p className="flex items-center gap-1 text-xs text-text-secondary">
            <Users className="h-3 w-3" strokeWidth={1.75} />
            {campaign.totals.contributor_count} {campaign.unit}
          </p>
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
  const [category, setCategory] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const { campaigns, isLoading, hasNextPage, isFetchingNextPage, fetchNextPage } =
    useExploreCampaigns(debouncedSearch, category)
  const { data: categories } = useCampaignCategories()
  const { data: featured } = useFeaturedCampaign()
  const hasFilters = Boolean(debouncedSearch) || Boolean(category)

  return (
    <div className="min-h-screen bg-bg-subtle">
      <header className="mx-auto flex max-w-5xl items-center justify-between px-6 py-5">
        <Link to="/" className="font-display text-lg font-bold tracking-tight">
          Juntalo
        </Link>
        <div className="flex items-center gap-3">
          <Link
            to="/como-funciona"
            className="text-sm font-medium text-text-secondary hover:text-text-primary"
          >
            Cómo funciona
          </Link>
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

        {!hasFilters && featured && <FeaturedCampaignCard campaign={featured} />}

        <div className="relative mb-4 max-w-md">
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

        {categories && categories.length > 0 && (
          <div className="mb-6 flex flex-wrap gap-2">
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
        )}

        {isLoading && <p className="text-text-secondary">Cargando…</p>}

        {!isLoading && campaigns.length === 0 && (
          <p className="text-text-secondary">
            {hasFilters
              ? 'No hay campañas activas que coincidan con el filtro.'
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
