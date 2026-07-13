import { Card } from '../../../shared/ui'
import { useContributionStatus } from '../hooks/useContribute'

interface ContributeSuccessPanelProps {
  contributionId: string
  initialStatus: string
  onClose: () => void
}

const statusCopy: Record<string, { title: string; description: string }> = {
  pending: { title: 'Confirmando tu aporte…', description: 'Esto toma solo un momento.' },
  confirmed: { title: '¡Gracias por tu aporte!', description: 'Tu pago fue confirmado.' },
  failed: { title: 'El pago no se pudo procesar', description: 'Intenta nuevamente.' },
}

// Pantalla de confirmación con polling (Etapa 4 §4): el flujo de pago del
// mock es asíncrono vía webhook, así que el estado real llega poco después.
export function ContributeSuccessPanel({ contributionId, initialStatus, onClose }: ContributeSuccessPanelProps) {
  const { data } = useContributionStatus(contributionId, initialStatus)
  const status = data?.status ?? initialStatus
  const copy = statusCopy[status] ?? statusCopy.pending

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <Card className="w-full max-w-sm text-center">
        <h2 className="mb-2 text-lg font-semibold">{copy.title}</h2>
        <p className="mb-4 text-sm text-text-secondary">{copy.description}</p>
        {status === 'pending' && (
          <div className="mb-4 h-1 w-full overflow-hidden rounded-full bg-bg-subtle">
            <div className="h-full w-1/3 animate-pulse rounded-full bg-brand" />
          </div>
        )}
        <button onClick={onClose} className="text-sm text-brand hover:underline">
          Cerrar
        </button>
      </Card>
    </div>
  )
}
