import { useState, type ChangeEvent, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Button, Input } from '../../../shared/ui'
import { errorMessage } from '../../../shared/api/errors'
import { formatCLP } from '../../../shared/lib/clp'
import { useContribute } from '../hooks/useContribute'
import type { StartContributionResult } from '../api'

interface ContributeSheetProps {
  slug: string
  cta: string
  onClose: () => void
  onSuccess: (result: StartContributionResult) => void
  // Rifa: precio fijo por número (no monto libre) y cuántos quedan — si
  // raffleAvailable llega en 0, el formulario se deshabilita en vez de
  // dejar que el usuario intente comprar un número que ya no existe.
  raffleUnitPrice?: number
  raffleAvailable?: number
}

// Bottom-sheet: el aporte ocurre sin salir de la página, clave en el in-app
// browser de WhatsApp (Etapa 5 §3).
export function ContributeSheet({
  slug,
  cta,
  onClose,
  onSuccess,
  raffleUnitPrice,
  raffleAvailable,
}: ContributeSheetProps) {
  const { mutateAsync, isPending } = useContribute(slug)
  const isRaffle = raffleUnitPrice !== undefined
  const soldOut = isRaffle && raffleAvailable === 0
  const [fullName, setFullName] = useState('')
  // amount guarda solo dígitos (fuente de verdad); el input muestra el
  // monto formateado como CLP ($7.777) mientras se escribe. En una rifa el
  // precio es fijo, así que arranca ya con ese valor y no se puede editar.
  const [amount, setAmount] = useState(isRaffle ? String(raffleUnitPrice) : '')
  const [isAnonymous, setIsAnonymous] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState<string | null>(null)

  function handleAmountChange(e: ChangeEvent<HTMLInputElement>) {
    setAmount(e.target.value.replace(/\D/g, ''))
  }

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
      // Con una pasarela real (Webpay) el pago no se resuelve en la página:
      // hay que sacar al aportante entero del sitio para que entre su
      // tarjeta en Transbank. El mock nunca manda redirect_url, así que este
      // camino no cambia nada para el flujo simulado.
      if (result.payment.redirect_url) {
        window.location.href = result.payment.redirect_url
        return
      }
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
          {isRaffle ? (
            <div className="rounded-lg border border-border-default bg-bg-subtle px-3 py-2 text-sm">
              <span className="text-text-secondary">Precio por número: </span>
              <span className="font-semibold text-text-primary">{formatCLP(raffleUnitPrice!)}</span>
              {raffleAvailable !== undefined && (
                <p className="mt-0.5 text-xs text-text-secondary">
                  {soldOut ? 'No quedan números disponibles' : `Quedan ${raffleAvailable} números disponibles`}
                </p>
              )}
            </div>
          ) : (
            <Input
              type="text"
              inputMode="numeric"
              placeholder="Monto en CLP"
              value={amount ? formatCLP(Number(amount)) : ''}
              onChange={handleAmountChange}
              required
            />
          )}
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
            <Button type="submit" disabled={isPending || soldOut} className="flex-1">
              {isPending ? 'Procesando…' : soldOut ? 'Sin números disponibles' : cta}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
