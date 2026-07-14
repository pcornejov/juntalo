import { Link } from 'react-router-dom'
import { Users } from 'lucide-react'
import { Card, Progress } from '../../../shared/ui'
import { formatCLP } from '../../../shared/lib/clp'
import { StatusBadge } from './StatusBadge'
import { typeLabels } from '../typeMeta'
import type { Campaign } from '../api'

export function CampaignCard({ campaign }: { campaign: Campaign }) {
  return (
    <Link to={`/dashboard/campaigns/${campaign.id}`}>
      <Card className="h-full transition-shadow hover:shadow-md">
        <div className="mb-2 flex items-start justify-between gap-2">
          <div>
            <h3 className="font-display text-sm font-bold">{campaign.title}</h3>
            <p className="text-[11px] text-text-secondary">
              {typeLabels[campaign.type_key] ?? campaign.type_key}
            </p>
          </div>
          <StatusBadge status={campaign.status} />
        </div>
        <p className="mb-3 text-sm text-text-secondary">
          <span className="font-semibold text-brand-hover">{formatCLP(campaign.totals.raised_gross)}</span>
          {campaign.goal_amount ? ` de ${formatCLP(campaign.goal_amount)}` : ' recaudados'}
        </p>
        {campaign.goal_amount && (
          <Progress value={campaign.totals.raised_gross} max={campaign.goal_amount} />
        )}
        <p className="mt-3 flex items-center gap-1 text-xs text-text-secondary">
          <Users className="h-3 w-3" strokeWidth={1.75} />
          {campaign.totals.contributor_count} participantes
        </p>
      </Card>
    </Link>
  )
}
