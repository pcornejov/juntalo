import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import * as authApi from '../api'
import { errorMessage } from '../../../shared/api/errors'
import { Button, Card, Input } from '../../../shared/ui'

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''
  const navigate = useNavigate()
  const [newPassword, setNewPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [done, setDone] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)
    try {
      await authApi.resetPassword(token, newPassword)
      setDone(true)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-sm flex-col justify-center p-6">
      <Link
        to="/"
        className="mb-6 self-start text-sm font-medium text-text-secondary hover:text-text-primary"
      >
        ← Juntalo
      </Link>
      <h1 className="mb-6 text-2xl font-semibold">Nueva contraseña</h1>
      <Card>
        {!token ? (
          <p className="text-sm text-danger">
            Este link no es válido. Pide uno nuevo desde{' '}
            <Link to="/forgot-password" className="underline">
              recuperar contraseña
            </Link>
            .
          </p>
        ) : done ? (
          <div className="space-y-3 text-center">
            <p className="text-sm text-text-primary">Tu contraseña fue actualizada.</p>
            <Button className="w-full" onClick={() => navigate('/login')}>
              Ir a iniciar sesión
            </Button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              type="password"
              placeholder="Contraseña nueva (mínimo 8 caracteres)"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              minLength={8}
              autoComplete="new-password"
            />
            {error && <p className="text-sm text-danger">{error}</p>}
            <Button type="submit" disabled={isSubmitting} className="w-full">
              {isSubmitting ? 'Guardando…' : 'Guardar contraseña'}
            </Button>
          </form>
        )}
      </Card>
    </div>
  )
}
