import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Link2, Plus, Search } from 'lucide-react'
import { useCampaigns } from '../hooks/useCampaigns'
import { useAuth } from '../../auth/hooks/useAuth'
import { Button, Input } from '../../../shared/ui'
import { CampaignCard } from '../components/CampaignCard'
import { normalizeForSearch } from '../../../shared/lib/search'
import type { Campaign } from '../api'

const statusOptions: { value: Campaign['status'] | 'all'; label: string }[] = [
  { value: 'all', label: 'Todos los estados' },
  { value: 'draft', label: 'Borrador' },
  { value: 'active', label: 'Activa' },
  { value: 'paused', label: 'Pausada' },
  { value: 'finished', label: 'Finalizada' },
  { value: 'suspended', label: 'Suspendida' },
]

export function CampaignListPage() {
  const { campaigns, isLoading, hasNextPage, isFetchingNextPage, fetchNextPage } = useCampaigns()
  const { organization } = useAuth()
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<Campaign['status'] | 'all'>('all')
  const [copied, setCopied] = useState(false)

  async function copyOrgLink() {
    if (!organization) return
    const url = `${window.location.origin}/org/${organization.slug}`
    await navigator.clipboard.writeText(url)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const filtered = useMemo(() => {
    const query = normalizeForSearch(search.trim())
    return campaigns.filter((c) => {
      if (status !== 'all' && c.status !== status) return false
      if (query && !normalizeForSearch(c.title).includes(query)) return false
      return true
    })
  }, [campaigns, search, status])

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="font-display text-2xl font-bold tracking-tight">Mis campañas</h1>
        <div className="flex items-center gap-2">
          {organization && (
            <Button variant="secondary" className="flex items-center gap-1.5" onClick={copyOrgLink}>
              <Link2 className="h-4 w-4" strokeWidth={2} />
              {copied ? 'Link copiado' : 'Compartir mi página'}
            </Button>
          )}
          <Link to="/dashboard/campaigns/new">
            <Button className="flex items-center gap-1.5">
              <Plus className="h-4 w-4" strokeWidth={2} />
              Nueva campaña
            </Button>
          </Link>
        </div>
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

      {!isLoading && campaigns.length > 0 && (
        <div className="mb-4 flex flex-wrap items-center gap-2">
          <div className="relative flex-1 min-w-[200px]">
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
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value as Campaign['status'] | 'all')}
            className="rounded-lg border border-border-default bg-bg-surface px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
          >
            {statusOptions.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </div>
      )}

      {!isLoading && campaigns.length > 0 && filtered.length === 0 && (
        <p className="text-text-secondary">No hay campañas que coincidan con la búsqueda.</p>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {filtered.map((c) => (
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
