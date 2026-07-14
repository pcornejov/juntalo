import { Link } from 'react-router-dom'
import { ShieldCheck, Users } from 'lucide-react'
import { formatCLP } from '../../../shared/lib/clp'
import { Progress } from '../../../shared/ui'
import type { PublicCampaign } from '../api'

export function VerifiedBadge() {
  return (
    <span className="flex items-center gap-1 rounded-full bg-accent-tint px-2 py-0.5 text-[10px] font-bold text-brand-hover">
      <ShieldCheck className="h-3 w-3" strokeWidth={1.75} />
      Verificada
    </span>
  )
}

export function CampaignExploreCard({ campaign }: { campaign: PublicCampaign }) {
  return (
    <Link to={`/public/${campaign.slug}`}>
      <div className="h-full overflow-hidden rounded-xl border border-border-default bg-bg-surface shadow-sm transition-shadow hover:shadow-md">
        <div className="h-32 w-full bg-bg-subtle">
          {campaign.cover_url && (
            <img src={campaign.cover_url} alt="" className="h-32 w-full object-cover" />
          )}
        </div>
        <div className="space-y-3 p-4">
          <div className="flex items-start justify-between gap-2">
            <h3 className="font-display text-sm font-bold leading-tight">{campaign.title}</h3>
            {campaign.is_verified && <VerifiedBadge />}
          </div>
          <p className="text-sm text-text-secondary">
            <span className="font-semibold text-brand-hover">
              {formatCLP(campaign.totals.raised_gross)}
            </span>
            {campaign.goal_amount ? ` de ${formatCLP(campaign.goal_amount)}` : ' recaudados'}
          </p>
          {campaign.goal_amount && (
            <Progress value={campaign.totals.raised_gross} max={campaign.goal_amount} />
          )}
          <div className="flex items-center justify-between">
            <p className="flex items-center gap-1 text-xs text-text-secondary">
              <Users className="h-3 w-3" strokeWidth={1.75} />
              {campaign.totals.contributor_count} {campaign.unit}
            </p>
            <span className="text-xs font-semibold text-brand-hover">{campaign.cta} →</span>
          </div>
        </div>
      </div>
    </Link>
  )
}
