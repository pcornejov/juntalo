import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Button, Input } from '../../../shared/ui'
import { errorMessage } from '../../../shared/api/errors'
import { useContribute } from '../hooks/useContribute'
import type { StartContributionResult } from '../api'

interface ContributeSheetProps {
  slug: string
  cta: string
  onClose: () => void
  onSuccess: (result: StartContributionResult) => void
}

// Bottom-sheet: el aporte ocurre sin salir de la página, clave en el in-app
// browser de WhatsApp (Etapa 5 §3).
export function ContributeSheet({ slug, cta, onClose, onSuccess }: ContributeSheetProps) {
  const { mutateAsync, isPending } = useContribute(slug)
  const [fullName, setFullName] = useState('')
  const [amount, setAmount] = useState('')
  const [isAnonymous, setIsAnonymous] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    try {
      const result = await mutateAsync({
        full_name: fullName,
        amount: Number(amount),
        is_anonymous: isAnonymous,
        message: message || undefined,
      })
      onSuccess(result)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end sm:items-center sm:justify-center">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <div className="relative w-full max-w-md rounded-t-2xl bg-bg-surface p-6 sm:rounded-2xl">
        <h2 className="mb-4 text-lg font-semibold">{cta}</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            placeholder="Tu nombre"
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
            required
          />
          <Input
            type="number"
            min={1}
            placeholder="Monto en CLP"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            required
          />
          <textarea
            className="w-full rounded-lg border border-border-default bg-bg-surface px-3 py-2 text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
            placeholder="Mensaje (opcional)"
            rows={2}
            value={message}
            onChange={(e) => setMessage(e.target.value)}
          />
          <label className="flex items-center gap-2 text-sm text-text-secondary">
            <input
              type="checkbox"
              checked={isAnonymous}
              onChange={(e) => setIsAnonymous(e.target.checked)}
            />
            Aportar de forma anónima
          </label>
          {error && <p className="text-sm text-danger">{error}</p>}
          <p className="text-xs text-text-secondary">
            Tus datos se usan solo para procesar este aporte y contactarte si es necesario. No
            los compartimos con terceros. Ver{' '}
            <Link to="/privacidad" className="underline" target="_blank" rel="noopener noreferrer">
              política de privacidad
            </Link>
            .
          </p>
          <div className="flex gap-2">
            <Button type="button" variant="secondary" onClick={onClose} className="flex-1">
              Cancelar
            </Button>
            <Button type="submit" disabled={isPending} className="flex-1">
              {isPending ? 'Procesando…' : cta}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
