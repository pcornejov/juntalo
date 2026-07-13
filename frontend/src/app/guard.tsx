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
