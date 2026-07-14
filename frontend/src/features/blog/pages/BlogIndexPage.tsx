import { Link } from 'react-router-dom'
import { BookOpen } from 'lucide-react'
import { blogArticles } from '../articles'
import { Footer, ThemeToggle } from '../../../shared/ui'

// Índice del contenido educativo estático (inspirado en el blog de Ceneka):
// consejos prácticos para quien va a crear su primera campaña.
export function BlogIndexPage() {
  return (
    <div className="min-h-screen bg-bg-subtle">
      <header className="mx-auto flex max-w-3xl items-center justify-between px-6 py-5">
        <Link to="/" className="font-display text-lg font-bold tracking-tight">
          Juntalo
        </Link>
        <div className="flex items-center gap-3">
          <Link to="/explorar" className="text-sm font-medium text-text-secondary hover:text-text-primary">
            Explorar campañas
          </Link>
          <ThemeToggle />
        </div>
      </header>

      <div className="mx-auto max-w-3xl px-6 pb-16">
        <h1 className="mb-2 flex items-center gap-2 font-display text-2xl font-bold tracking-tight">
          <BookOpen className="h-6 w-6 text-brand-hover" strokeWidth={1.75} />
          Consejos para tu campaña
        </h1>
        <p className="mb-10 text-text-secondary">
          Ideas prácticas para escribir, compartir y administrar tu campaña.
        </p>

        <div className="space-y-4">
          {blogArticles.map((a) => (
            <Link
              key={a.slug}
              to={`/consejos/${a.slug}`}
              className="block rounded-xl border border-border-default bg-bg-surface p-4 transition-shadow hover:shadow-md"
            >
              <h2 className="mb-1 font-display text-base font-bold">{a.title}</h2>
              <p className="text-sm text-text-secondary">{a.excerpt}</p>
            </Link>
          ))}
        </div>
      </div>

      <Footer />
    </div>
  )
}
