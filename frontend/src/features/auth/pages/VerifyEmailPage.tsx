import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import * as authApi from '../api'
import { useAuth } from '../hooks/useAuth'
import { errorMessage } from '../../../shared/api/errors'
import { Button, Card } from '../../../shared/ui'

type Status = 'verifying' | 'success' | 'error'

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''
  const { user, refreshUser } = useAuth()
  const [status, setStatus] = useState<Status>(token ? 'verifying' : 'error')
  const [error, setError] = useState<string | null>(token ? null : 'Este link no es válido.')

  useEffect(() => {
    if (!token) return
    let cancelled = false
    authApi
      .verifyEmail(token)
      .then(async () => {
        if (cancelled) return
        // Si ya había sesión iniciada, refresca /me para que el banner
        // "verifica tu email" desaparezca sin pedir reloguear.
        if (user) await refreshUser().catch(() => undefined)
        if (!cancelled) setStatus('success')
      })
      .catch((err) => {
        if (!cancelled) {
          setStatus('error')
          setError(errorMessage(err))
        }
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token])

  return (
    <div className="mx-auto flex min-h-screen max-w-sm flex-col justify-center p-6">
      <Link
        to="/"
        className="mb-6 self-start text-sm font-medium text-text-secondary hover:text-text-primary"
      >
        ← Juntalo
      </Link>
      <h1 className="mb-6 text-2xl font-semibold">Verificar email</h1>
      <Card className="text-center">
        {status === 'verifying' && <p className="text-sm text-text-secondary">Verificando…</p>}
        {status === 'success' && (
          <div className="space-y-3">
            <p className="text-sm text-text-primary">¡Tu email quedó verificado!</p>
            <Link to={user ? '/dashboard' : '/login'}>
              <Button className="w-full">{user ? 'Ir al dashboard' : 'Iniciar sesión'}</Button>
            </Link>
          </div>
        )}
        {status === 'error' && <p className="text-sm text-danger">{error}</p>}
      </Card>
    </div>
  )
}
