import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { errorMessage } from '../../../shared/api/errors'
import { GoogleButton, isGoogleLoginEnabled } from '../components/GoogleButton'
import { Button, Card, Input } from '../../../shared/ui'

export function LoginPage() {
  const { login, loginWithGoogle } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)
    try {
      await login(email, password)
      navigate('/dashboard')
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleGoogleCredential(credential: string) {
    setError(null)
    try {
      await loginWithGoogle(credential)
      navigate('/dashboard')
    } catch (err) {
      setError(errorMessage(err))
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
      <h1 className="mb-6 text-2xl font-semibold">Iniciar sesión</h1>
      <Card>
        {isGoogleLoginEnabled() && (
          <div className="mb-4 space-y-4">
            <GoogleButton onCredential={handleGoogleCredential} />
            <div className="flex items-center gap-3 text-xs text-text-secondary">
              <div className="h-px flex-1 bg-border-default" />
              o
              <div className="h-px flex-1 bg-border-default" />
            </div>
          </div>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
          <Input
            type="password"
            placeholder="Contraseña"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="current-password"
          />
          {error && <p className="text-sm text-danger">{error}</p>}
          <div className="text-right">
            <Link to="/forgot-password" className="text-xs text-text-secondary hover:underline">
              ¿Olvidaste tu contraseña?
            </Link>
          </div>
          <Button type="submit" disabled={isSubmitting} className="w-full">
            {isSubmitting ? 'Ingresando…' : 'Ingresar'}
          </Button>
        </form>
      </Card>
      <p className="mt-4 text-center text-sm text-text-secondary">
        ¿No tienes cuenta?{' '}
        <Link to="/register" className="text-brand hover:underline">
          Crea una
        </Link>
      </p>
    </div>
  )
}
