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

export interface StartContributionInput {
  full_name: string
  email?: string
  phone?: string
  amount: number
  is_anonymous?: boolean
  message?: string
}

export interface StartContributionResult {
  contribution_id: string
  payment: {
    status: 'pending' | 'confirmed' | 'failed'
    redirect_url?: string
  }
}

export function startContribution(slug: string, idempotencyKey: string, input: StartContributionInput) {
  return apiClient.post<StartContributionResult>(`/public/campaigns/${slug}/contributions`, input, {
    skipAuth: true,
    headers: { 'Idempotency-Key': idempotencyKey },
  })
}

export function getContributionStatus(id: string) {
  return apiClient.get<{ status: string }>(`/public/contributions/${id}/status`)
}
