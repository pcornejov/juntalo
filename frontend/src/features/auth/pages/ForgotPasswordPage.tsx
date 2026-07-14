import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import * as authApi from '../api'
import { errorMessage } from '../../../shared/api/errors'
import { Button, Card, Input } from '../../../shared/ui'

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [message, setMessage] = useState<string | null>(null)
  // Atajo de este deploy de prueba: no hay envío de email real todavía, así
  // que el backend puede devolver el link directo (EXPOSE_RESET_LINKS=true).
  // En un despliegue real este campo nunca llega y el usuario revisa su correo.
  const [devResetLink, setDevResetLink] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)
    try {
      const res = await authApi.forgotPassword(email)
      setMessage(res.message)
      setDevResetLink(res.reset_token ? `${window.location.origin}/reset-password?token=${res.reset_token}` : null)
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
      <h1 className="mb-6 text-2xl font-semibold">Recuperar contraseña</h1>
      <Card>
        {message ? (
          <div className="space-y-3">
            <p className="text-sm text-text-primary">{message}</p>
            {devResetLink && (
              <div className="rounded-lg border border-dashed border-border-default p-3">
                <p className="mb-1 text-xs font-semibold text-text-secondary">
                  Modo de prueba (sin envío de email real):
                </p>
                <a href={devResetLink} className="break-all text-xs text-brand underline">
                  {devResetLink}
                </a>
              </div>
            )}
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <p className="text-sm text-text-secondary">
              Ingresa tu email y te enviaremos instrucciones para restablecer tu contraseña.
            </p>
            <Input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="email"
            />
            {error && <p className="text-sm text-danger">{error}</p>}
            <Button type="submit" disabled={isSubmitting} className="w-full">
              {isSubmitting ? 'Enviando…' : 'Enviar instrucciones'}
            </Button>
          </form>
        )}
      </Card>
      <p className="mt-4 text-center text-sm text-text-secondary">
        <Link to="/login" className="text-brand hover:underline">
          Volver a iniciar sesión
        </Link>
      </p>
    </div>
  )
}
