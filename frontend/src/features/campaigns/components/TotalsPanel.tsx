import { Card } from '../../../shared/ui'
import { formatCLP } from '../../../shared/lib/clp'
import type { Campaign } from '../api'

export function TotalsPanel({ campaign }: { campaign: Campaign }) {
  // El costo de servicio es la diferencia entre bruto y neto — no requiere
  // conocer la tasa de comisión, ya se refleja en lo que ya resta la vista
  // neta (que a su vez ya descuenta reembolsos, ver campaign_totals).
  const serviceCost = campaign.totals.raised_gross - campaign.totals.raised_net_approx

  return (
    <Card className="grid grid-cols-2 gap-4 sm:grid-cols-4">
      <div>
        <p className="text-xs uppercase tracking-wide text-text-secondary">Recaudado (bruto)</p>
        <p className="font-display text-lg font-bold tabular-nums text-brand-hover">
          {formatCLP(campaign.totals.raised_gross)}
        </p>
      </div>
      <div>
        <p className="text-xs uppercase tracking-wide text-text-secondary">Costo de servicio</p>
        <p className="font-display text-lg font-bold tabular-nums text-text-secondary">
          -{formatCLP(serviceCost)}
        </p>
      </div>
      <div>
        <p className="text-xs uppercase tracking-wide text-text-secondary">Recaudado (neto)</p>
        <p className="font-display text-lg font-bold tabular-nums">
          {formatCLP(campaign.totals.raised_net_approx)}
        </p>
      </div>
      <div>
        <p className="text-xs uppercase tracking-wide text-text-secondary">Participantes</p>
        <p className="font-display text-lg font-bold tabular-nums">{campaign.totals.contributor_count}</p>
      </div>
    </Card>
  )
}
