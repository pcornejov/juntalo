import { Link, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { getBlogArticle } from '../articles'
import { Footer, ThemeToggle } from '../../../shared/ui'

export function BlogArticlePage() {
  const { slug } = useParams<{ slug: string }>()
  const article = slug ? getBlogArticle(slug) : undefined

  return (
    <div className="min-h-screen bg-bg-subtle">
      <header className="mx-auto flex max-w-3xl items-center justify-between px-6 py-5">
        <Link to="/" className="font-display text-lg font-bold tracking-tight">
          Juntalo
        </Link>
        <ThemeToggle />
      </header>

      <div className="mx-auto max-w-3xl px-6 pb-16">
        <Link
          to="/consejos"
          className="mb-6 inline-flex items-center gap-1 text-sm font-medium text-text-secondary hover:text-text-primary"
        >
          <ArrowLeft className="h-4 w-4" strokeWidth={1.75} />
          Consejos
        </Link>

        {!article && <p className="text-text-secondary">Este artículo no existe.</p>}

        {article && (
          <>
            <h1 className="mb-6 font-display text-2xl font-bold tracking-tight">{article.title}</h1>
            <div className="space-y-4 text-text-primary">
              {article.body.map((paragraph, i) => (
                <p key={i} className="leading-relaxed">
                  {paragraph}
                </p>
              ))}
            </div>
          </>
        )}
      </div>

      <Footer />
    </div>
  )
}
