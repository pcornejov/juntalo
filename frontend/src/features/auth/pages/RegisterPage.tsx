import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { errorMessage } from '../../../shared/api/errors'
import { isCaptchaEnabled, TurnstileWidget } from '../components/TurnstileWidget'
import { GoogleButton, isGoogleLoginEnabled } from '../components/GoogleButton'
import { Button, Card, Input } from '../../../shared/ui'

export function RegisterPage() {
  const { register, loginWithGoogle } = useAuth()
  const navigate = useNavigate()
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [captchaToken, setCaptchaToken] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const captchaRequired = isCaptchaEnabled()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)
    try {
      await register(email, fullName, password, captchaToken || undefined)
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
      <h1 className="mb-6 text-2xl font-semibold">Crear cuenta</h1>
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
            placeholder="Nombre completo"
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
            required
            autoComplete="name"
          />
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
            placeholder="Contraseña (mínimo 8 caracteres)"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            autoComplete="new-password"
          />
          <TurnstileWidget onVerify={setCaptchaToken} />
          {error && <p className="text-sm text-danger">{error}</p>}
          <Button
            type="submit"
            disabled={isSubmitting || (captchaRequired && !captchaToken)}
            className="w-full"
          >
            {isSubmitting ? 'Creando…' : 'Crear cuenta'}
          </Button>
        </form>
      </Card>
      <p className="mt-4 text-center text-sm text-text-secondary">
        ¿Ya tienes cuenta?{' '}
        <Link to="/login" className="text-brand hover:underline">
          Inicia sesión
        </Link>
      </p>
    </div>
  )
}
