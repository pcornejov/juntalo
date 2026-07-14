import { Card } from '../../../shared/ui'
import { formatCLP } from '../../../shared/lib/clp'
import type { Campaign } from '../api'

export function TotalsPanel({ campaign }: { campaign: Campaign }) {
  return (
    <Card className="grid grid-cols-2 gap-4 sm:grid-cols-3">
      <div>
        <p className="text-xs uppercase tracking-wide text-text-secondary">Recaudado (bruto)</p>
        <p className="font-display text-lg font-bold tabular-nums text-brand-hover">
          {formatCLP(campaign.totals.raised_gross)}
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
