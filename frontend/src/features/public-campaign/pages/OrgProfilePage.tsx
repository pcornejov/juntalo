import { useParams, Link } from 'react-router-dom'
import { Users } from 'lucide-react'
import { useOrgProfile } from '../hooks/usePublicCampaign'
import { Button, Footer, ThemeToggle } from '../../../shared/ui'
import { CampaignExploreCard, VerifiedBadge } from '../components/CampaignExploreCard'

// Página pública persistente del organizador (/org/:slug) — inspirada en el
// link único de por vida de Ceneka: a diferencia de compartir una campaña
// puntual, un organizador recurrente (ONG, junta de vecinos) puede compartir
// un solo link fijo que lista todas sus campañas pasadas y activas.
export function OrgProfilePage() {
  const { slug } = useParams<{ slug: string }>()
  const { campaigns, name, isVerified, isLoading, hasNextPage, isFetchingNextPage, fetchNextPage } =
    useOrgProfile(slug)

  return (
    <div className="min-h-screen bg-bg-subtle">
      <header className="mx-auto flex max-w-5xl items-center justify-between px-6 py-5">
        <Link to="/" className="font-display text-lg font-bold tracking-tight">
          Juntalo
        </Link>
        <div className="flex items-center gap-3">
          <Link
            to="/explorar"
            className="text-sm font-medium text-text-secondary hover:text-text-primary"
          >
            Explorar
          </Link>
          <ThemeToggle />
        </div>
      </header>

      <div className="mx-auto max-w-5xl px-6 pb-16">
        {isLoading && <p className="text-text-secondary">Cargando…</p>}

        {!isLoading && !name && (
          <p className="text-text-secondary">Esta página no existe.</p>
        )}

        {name && (
          <>
            <div className="mb-8 flex items-center gap-4">
              <div className="flex h-16 w-16 flex-none items-center justify-center rounded-full bg-accent-tint font-display text-xl font-bold text-brand-hover">
                {name.trim().charAt(0).toUpperCase()}
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h1 className="font-display text-xl font-bold tracking-tight">{name}</h1>
                  {isVerified && <VerifiedBadge />}
                </div>
                <p className="flex items-center gap-1 text-sm text-text-secondary">
                  <Users className="h-3.5 w-3.5" strokeWidth={1.75} />
                  {campaigns.length} campaña{campaigns.length === 1 ? '' : 's'}
                </p>
              </div>
            </div>

            {campaigns.length === 0 && (
              <p className="text-text-secondary">Este organizador todavía no tiene campañas públicas.</p>
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
          </>
        )}
      </div>

      <Footer />
    </div>
  )
}
