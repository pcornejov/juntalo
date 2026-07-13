import { Outlet } from 'react-router-dom'
import { useAuth } from '../features/auth/hooks/useAuth'
import { Button } from '../shared/ui'

export function DashboardLayout() {
  const { user, organization, logout } = useAuth()

  return (
    <div className="min-h-screen">
      <header className="flex items-center justify-between border-b border-border-default bg-bg-surface px-6 py-4">
        <div>
          <p className="font-semibold">Juntalo</p>
          <p className="text-sm text-text-secondary">{organization?.name}</p>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm text-text-secondary">{user?.full_name}</span>
          <Button variant="secondary" onClick={() => logout()}>
            Cerrar sesión
          </Button>
        </div>
      </header>
      <main className="p-6">
        <Outlet />
      </main>
    </div>
  )
}
