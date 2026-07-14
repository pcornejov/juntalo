import { Fragment, useState } from 'react'
import { formatCLP } from '../../../shared/lib/clp'
import { Badge, Button, Input } from '../../../shared/ui'
import type { Participant } from '../api'
import { useRefundContribution } from '../hooks/useCampaigns'

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

function isRefundable(p: Participant) {
  return p.amount - p.refunded_amount > 0 && (p.status === 'confirmed' || p.status === 'refunded')
}

function RefundForm({ campaignId, p, onDone }: { campaignId: string; p: Participant; onDone: () => void }) {
  const remaining = p.amount - p.refunded_amount
  const [amount, setAmount] = useState(String(remaining))
  const [reason, setReason] = useState('')
  const refund = useRefundContribution(campaignId)

  const submit = () => {
    const parsed = Number(amount)
    if (!parsed || parsed <= 0 || parsed > remaining) return
    refund.mutate(
      { contributionId: p.contribution_id, amount: parsed, reason },
      { onSuccess: onDone },
    )
  }

  return (
    <div className="flex flex-wrap items-center gap-2 py-2">
      <Input
        type="number"
        min={1}
        max={remaining}
        value={amount}
        onChange={(e) => setAmount(e.target.value)}
        className="w-32"
        aria-label="Monto a reembolsar"
      />
      <Input
        type="text"
        placeholder="Motivo (opcional)"
        value={reason}
        onChange={(e) => setReason(e.target.value)}
        className="w-48"
      />
      <Button variant="danger" onClick={submit} disabled={refund.isPending}>
        {refund.isPending ? 'Reembolsando…' : 'Confirmar reembolso'}
      </Button>
      <Button variant="secondary" onClick={onDone} disabled={refund.isPending}>
        Cancelar
      </Button>
      {refund.isError && (
        <span className="text-xs text-danger">No se pudo procesar el reembolso.</span>
      )}
    </div>
  )
}

export function ParticipantsTable({ items, campaignId }: { items: Participant[]; campaignId: string }) {
  const [refunding, setRefunding] = useState<string | null>(null)

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
            <th className="py-2 pr-4"></th>
          </tr>
        </thead>
        <tbody>
          {items.map((p) => (
            <Fragment key={p.contribution_id}>
              <tr className="border-b border-border-default last:border-0">
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
                <td className="py-2 pr-4 text-right">
                  {isRefundable(p) && refunding !== p.contribution_id && (
                    <button
                      className="text-xs font-medium text-danger hover:underline"
                      onClick={() => setRefunding(p.contribution_id)}
                    >
                      Reembolsar
                    </button>
                  )}
                </td>
              </tr>
              {refunding === p.contribution_id && (
                <tr className="border-b border-border-default last:border-0 bg-bg-subtle">
                  <td colSpan={6} className="px-2">
                    <RefundForm campaignId={campaignId} p={p} onDone={() => setRefunding(null)} />
                  </td>
                </tr>
              )}
            </Fragment>
          ))}
        </tbody>
      </table>
    </div>
  )
}
