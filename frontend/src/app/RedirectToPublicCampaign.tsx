import { Navigate, useParams } from 'react-router-dom'

// El link "real" para compartir (/c/:slug) vive en el backend, que sirve OG
// tags para bots y redirige navegadores reales a /public/:slug (Etapa 4 §4).
// Esta ruta es una red de seguridad: si alguien abre /c/:slug directo en el
// dominio del frontend (link viejo cacheado, autocompletado, etc.), lo manda
// al mismo lugar en vez de mostrar un 404.
export function RedirectToPublicCampaign() {
  const { slug } = useParams<{ slug: string }>()
  return <Navigate to={`/public/${slug}`} replace />
}
