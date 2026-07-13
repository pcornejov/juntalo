import { Card } from '../../../shared/ui'
import { formatCLP } from '../../../shared/lib/clp'
import type { Campaign } from '../api'

export function TotalsPanel({ campaign }: { campaign: Campaign }) {
  return (
    <Card className="grid grid-cols-2 gap-4 sm:grid-cols-3">
      <div>
        <p className="text-xs text-text-secondary">Recaudado (bruto)</p>
        <p className="text-lg font-semibold">{formatCLP(campaign.totals.raised_gross)}</p>
      </div>
      <div>
        <p className="text-xs text-text-secondary">Recaudado (neto)</p>
        <p className="text-lg font-semibold">{formatCLP(campaign.totals.raised_net_approx)}</p>
      </div>
      <div>
        <p className="text-xs text-text-secondary">Participantes</p>
        <p className="text-lg font-semibold">{campaign.totals.contributor_count}</p>
      </div>
    </Card>
  )
}
