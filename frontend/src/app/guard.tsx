import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../features/auth/hooks/useAuth'

export function RequireAuth() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return <div className="p-6 text-text-secondary">Cargando…</div>
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}

// RequireAdmin manda al dashboard normal en vez de a un 403/404 propio — el
// gate real vive en el backend (middleware.RequireAdminUser, que responde
// 404 en cualquier request de admin de alguien fuera de la lista); acá solo
// evitamos mostrarle el link/página a quien no la va a poder usar igual.
export function RequireAdmin() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return <div className="p-6 text-text-secondary">Cargando…</div>
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  if (!user.is_admin) {
    return <Navigate to="/dashboard" replace />
  }

  return <Outlet />
}
