import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Check } from 'lucide-react'
import { useAuth } from '../../auth/hooks/useAuth'
import { errorMessage } from '../../../shared/api/errors'
import { Button, Card, Input } from '../../../shared/ui'
import { useUpdatePayoutInfo } from '../hooks/useOrganization'

const accountTypes = [
  { value: 'corriente', label: 'Cuenta corriente' },
  { value: 'vista', label: 'Cuenta vista' },
  { value: 'ahorro', label: 'Cuenta de ahorro' },
  { value: 'rut', label: 'Cuenta RUT' },
]

// Datos de transferencia para la liquidación manual (Etapa post-MVP): el
// operador de la plataforma transfiere a mano cada cierto tiempo — sin
// estos datos, no sabe a dónde.
export function PayoutSettingsPage() {
  const { organization, refreshUser } = useAuth()
  const update = useUpdatePayoutInfo()
  const [rut, setRut] = useState(organization?.rut ?? '')
  const [bank, setBank] = useState(organization?.payout_bank ?? '')
  const [accountType, setAccountType] = useState(organization?.payout_account_type ?? '')
  const [accountNumber, setAccountNumber] = useState(organization?.payout_account_number ?? '')
  const [holderName, setHolderName] = useState(organization?.payout_holder_name ?? '')
  const [error, setError] = useState<string | null>(null)
  const [saved, setSaved] = useState(false)

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSaved(false)
    update.mutate(
      {
        rut,
        payout_bank: bank,
        payout_account_type: accountType,
        payout_account_number: accountNumber,
        payout_holder_name: holderName,
      },
      {
        onSuccess: () => {
          setSaved(true)
          refreshUser()
        },
        onError: (err) => setError(errorMessage(err)),
      },
    )
  }

  return (
    <div className="mx-auto max-w-lg">
      <Link
        to="/dashboard"
        className="mb-4 inline-flex items-center gap-1 text-sm font-medium text-text-secondary hover:text-text-primary"
      >
        <ArrowLeft className="h-4 w-4" strokeWidth={1.75} />
        Volver
      </Link>
      <h1 className="mb-2 font-display text-2xl font-bold tracking-tight">Datos de transferencia</h1>
      <p className="mb-6 text-sm text-text-secondary">
        Así te transferimos lo recaudado — ver los detalles en las{' '}
        <Link to="/bases" className="underline" target="_blank" rel="noopener noreferrer">
          bases del sitio
        </Link>
        .
      </p>

      <Card>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm text-text-secondary">RUT</label>
            <Input value={rut} onChange={(e) => setRut(e.target.value)} placeholder="11.111.111-1" required />
          </div>
          <div>
            <label className="mb-1 block text-sm text-text-secondary">Nombre del titular</label>
            <Input value={holderName} onChange={(e) => setHolderName(e.target.value)} required />
          </div>
          <div>
            <label className="mb-1 block text-sm text-text-secondary">Banco</label>
            <Input value={bank} onChange={(e) => setBank(e.target.value)} placeholder="Ej: Banco Estado" required />
          </div>
          <div>
            <label className="mb-1 block text-sm text-text-secondary">Tipo de cuenta</label>
            <select
              value={accountType}
              onChange={(e) => setAccountType(e.target.value)}
              required
              className="w-full rounded-lg border border-border-default bg-bg-surface px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
            >
              <option value="" disabled>
                Selecciona un tipo
              </option>
              {accountTypes.map((t) => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm text-text-secondary">Número de cuenta</label>
            <Input
              value={accountNumber}
              onChange={(e) => setAccountNumber(e.target.value)}
              required
            />
          </div>
          {error && <p className="text-sm text-danger">{error}</p>}
          {saved && (
            <p className="flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400">
              <Check className="h-4 w-4" strokeWidth={2} />
              Datos guardados.
            </p>
          )}
          <Button type="submit" disabled={update.isPending} className="w-full">
            {update.isPending ? 'Guardando…' : 'Guardar datos'}
          </Button>
        </form>
      </Card>
    </div>
  )
}
