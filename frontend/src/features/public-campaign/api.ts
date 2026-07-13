import { apiClient } from '../../shared/api/client'
import type { Totals } from '../campaigns/api'

export interface PublicCampaign {
  title: string
  description: string
  cover_url?: string
  goal_amount?: number
  status: string
  totals: Totals
  cta: string
  unit: string
}

export function getPublicCampaign(slug: string) {
  return apiClient.get<PublicCampaign>(`/public/campaigns/${slug}`)
}
