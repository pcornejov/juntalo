import { Link } from 'react-router-dom'
import { Card, Progress } from '../../../shared/ui'
import { formatCLP } from '../../../shared/lib/clp'
import { StatusBadge } from './StatusBadge'
import type { Campaign } from '../api'

export function CampaignCard({ campaign }: { campaign: Campaign }) {
  return (
    <Link to={`/dashboard/campaigns/${campaign.id}`}>
      <Card className="h-full transition-shadow hover:shadow-md">
        <div className="mb-2 flex items-start justify-between gap-2">
          <h3 className="font-medium">{campaign.title}</h3>
          <StatusBadge status={campaign.status} />
        </div>
        <p className="mb-3 text-sm text-text-secondary">
          {formatCLP(campaign.totals.raised_gross)}
          {campaign.goal_amount ? ` de ${formatCLP(campaign.goal_amount)}` : ' recaudados'}
        </p>
        {campaign.goal_amount && (
          <Progress value={campaign.totals.raised_gross} max={campaign.goal_amount} />
        )}
        <p className="mt-3 text-xs text-text-secondary">
          {campaign.totals.contributor_count} participantes
        </p>
      </Card>
    </Link>
  )
}
