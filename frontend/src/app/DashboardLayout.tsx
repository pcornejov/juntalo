import { Link, Outlet } from 'react-router-dom'
import { Landmark, LogOut, ShieldCheck } from 'lucide-react'
import { useAuth } from '../features/auth/hooks/useAuth'
import { VerifyEmailBanner } from '../features/auth/components/VerifyEmailBanner'
import { Button, ThemeToggle } from '../shared/ui'

export function DashboardLayout() {
  const { user, organization, logout } = useAuth()

  return (
    <div className="min-h-screen">
      <header className="flex items-center justify-between border-b border-border-default bg-bg-surface px-6 py-4">
        <div>
          <Link to="/dashboard" className="font-display text-lg font-bold tracking-tight">
            Juntalo
          </Link>
          <p className="text-sm text-text-secondary">{organization?.name}</p>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to="/dashboard/payout"
            className="flex items-center gap-1.5 text-sm font-medium text-text-secondary hover:text-text-primary"
          >
            <Landmark className="h-3.5 w-3.5" strokeWidth={1.75} />
            Datos de transferencia
          </Link>
          {user?.is_admin && (
            <Link
              to="/dashboard/backoffice"
              className="flex items-center gap-1.5 text-sm font-medium text-text-secondary hover:text-text-primary"
            >
              <ShieldCheck className="h-3.5 w-3.5" strokeWidth={1.75} />
              Backoffice
            </Link>
          )}
          <ThemeToggle />
          <span className="hidden text-sm text-text-secondary sm:inline">{user?.full_name}</span>
          <Button
            variant="secondary"
            className="flex items-center gap-1.5"
            onClick={() => logout()}
          >
            <LogOut className="h-3.5 w-3.5" strokeWidth={1.75} />
            Cerrar sesión
          </Button>
        </div>
      </header>
      <main className="mx-auto max-w-5xl p-6">
        {user && !user.email_verified && <VerifyEmailBanner />}
        <Outlet />
      </main>
    </div>
  )
}
