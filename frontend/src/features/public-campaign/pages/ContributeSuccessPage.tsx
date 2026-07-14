import { CheckCircle2, XCircle, Loader2 } from 'lucide-react'
import { Card } from '../../../shared/ui'

interface ContributeSuccessPanelProps {
  status: string
  onClose: () => void
}

const statusCopy: Record<string, { title: string; description: string }> = {
  pending: { title: 'Confirmando tu aporte…', description: 'Esto toma solo un momento.' },
  confirmed: { title: '¡Gracias por tu aporte!', description: 'Tu pago fue confirmado.' },
  failed: { title: 'El pago no se pudo procesar', description: 'Intenta nuevamente.' },
}

// Tratamiento visual por estado — antes las tres pantallas se veían
// idénticas (mismo panel neutro), lo que hacía que un pago confirmado no se
// sintiera distinto a uno todavía pendiente. Verde+check para éxito, rojo
// para fallo, acento de marca para pendiente — colores directos de Tailwind
// (no --color-danger) para que el éxito lea "verde" de verdad y no el
// índigo de marca, que ya está copado por el resto de la UI.
const statusStyles: Record<string, { ring: string; icon: string; iconEl: typeof CheckCircle2 }> = {
  pending: { ring: 'bg-accent-tint', icon: 'text-brand-hover', iconEl: Loader2 },
  confirmed: { ring: 'bg-green-100 dark:bg-green-500/15', icon: 'text-green-600 dark:text-green-400', iconEl: CheckCircle2 },
  failed: { ring: 'bg-red-100 dark:bg-red-500/15', icon: 'text-red-600 dark:text-red-400', iconEl: XCircle },
}

// Pantalla de confirmación (Etapa 4 §4): el flujo de pago del mock/Webpay es
// asíncrono vía webhook, así que el estado real llega poco después. El
// polling vive en PublicCampaignPage para poder también reflejar el
// "actualizando…" en el total recaudado mientras el pago sigue pendiente.
export function ContributeSuccessPanel({ status, onClose }: ContributeSuccessPanelProps) {
  const copy = statusCopy[status] ?? statusCopy.pending
  const style = statusStyles[status] ?? statusStyles.pending
  const Icon = style.iconEl

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <Card className="w-full max-w-sm text-center">
        <div className={`mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full ${style.ring}`}>
          <Icon
            className={`h-8 w-8 ${style.icon} ${status === 'pending' ? 'animate-spin' : ''}`}
            strokeWidth={1.75}
          />
        </div>
        <h2 className="mb-2 font-display text-lg font-bold">{copy.title}</h2>
        <p className="mb-4 text-sm text-text-secondary">{copy.description}</p>
        <button onClick={onClose} className="text-sm font-medium text-brand-hover hover:underline">
          Cerrar
        </button>
      </Card>
    </div>
  )
}
