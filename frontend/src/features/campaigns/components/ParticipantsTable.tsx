import { formatCLP } from '../../../shared/lib/clp'
import { Badge } from '../../../shared/ui'
import type { Participant } from '../api'

const statusTone: Record<Participant['status'], 'neutral' | 'success' | 'warning' | 'danger'> = {
  pending: 'warning',
  confirmed: 'success',
  failed: 'danger',
  refunded: 'neutral',
}

const statusLabel: Record<Participant['status'], string> = {
  pending: 'Pendiente',
  confirmed: 'Confirmado',
  failed: 'Fallido',
  refunded: 'Reembolsado',
}

export function ParticipantsTable({ items }: { items: Participant[] }) {
  if (items.length === 0) {
    return <p className="text-sm text-text-secondary">Todavía no hay participantes.</p>
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-border-default text-text-secondary">
            <th className="py-2 pr-4">Nombre</th>
            <th className="py-2 pr-4">Contacto</th>
            <th className="py-2 pr-4">Monto</th>
            <th className="py-2 pr-4">Estado</th>
            <th className="py-2 pr-4">Fecha</th>
          </tr>
        </thead>
        <tbody>
          {items.map((p) => (
            <tr key={p.contribution_id} className="border-b border-border-default last:border-0">
              <td className="py-2 pr-4">
                {p.full_name}
                {p.is_anonymous && <span className="ml-1 text-xs text-text-secondary">(anónimo en público)</span>}
              </td>
              <td className="py-2 pr-4 text-text-secondary">{p.email || p.phone || '—'}</td>
              <td className="py-2 pr-4">
                {formatCLP(p.amount)}
                {p.refunded_amount > 0 && (
                  <span className="ml-1 text-xs text-danger">(-{formatCLP(p.refunded_amount)})</span>
                )}
              </td>
              <td className="py-2 pr-4">
                <Badge tone={statusTone[p.status]}>{statusLabel[p.status]}</Badge>
              </td>
              <td className="py-2 pr-4 text-text-secondary">
                {new Date(p.created_at).toLocaleDateString('es-CL')}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
