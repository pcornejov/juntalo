import { apiClient } from '../../shared/api/client'

export interface Totals {
  raised_gross: number
  raised_net_approx: number
  contributor_count: number
}

export interface Campaign {
  id: string
  type_key: string
  title: string
  slug: string
  description: string
  cover_url?: string
  goal_amount?: number
  status: 'draft' | 'active' | 'paused' | 'finished' | 'suspended'
  public_url: string
  totals: Totals
  created_at: string
}

export interface CampaignType {
  key: string
  name: string
  cta: string
  unit: string
  requires_goal_amount: boolean
  allows_free_amount: boolean
}

export function listCampaigns() {
  return apiClient.get<{ items: Campaign[] }>('/campaigns')
}

export function getCampaign(id: string) {
  return apiClient.get<Campaign>(`/campaigns/${id}`)
}

export function createCampaign(input: {
  type_key: string
  title: string
  description?: string
  goal_amount?: number
}) {
  return apiClient.post<Campaign>('/campaigns', input)
}

export function publishCampaign(id: string) {
  return apiClient.post<Campaign>(`/campaigns/${id}/publish`)
}

export function pauseCampaign(id: string) {
  return apiClient.post<Campaign>(`/campaigns/${id}/pause`)
}

export function resumeCampaign(id: string) {
  return apiClient.post<Campaign>(`/campaigns/${id}/resume`)
}

export function finishCampaign(id: string) {
  return apiClient.post<Campaign>(`/campaigns/${id}/finish`)
}

export function listCampaignTypes() {
  return apiClient.get<{ items: CampaignType[] }>('/meta/campaign-types')
}

export function uploadCoverImage(file: File) {
  const form = new FormData()
  form.append('file', file)
  form.append('kind', 'campaign_cover')
  return apiClient.upload<{ id: string; url: string }>('/files', form)
}

export interface Participant {
  contribution_id: string
  full_name: string
  email?: string
  phone?: string
  amount: number
  refunded_amount: number
  is_anonymous: boolean
  status: 'pending' | 'confirmed' | 'failed' | 'refunded'
  created_at: string
}

export function listParticipants(campaignId: string) {
  return apiClient.get<{ items: Participant[] }>(`/campaigns/${campaignId}/contributions`)
}

export function attachCover(campaignId: string, coverFileId: string, current: Campaign) {
  return apiClient.patch<Campaign>(`/campaigns/${campaignId}`, {
    title: current.title,
    description: current.description,
    goal_amount: current.goal_amount,
    cover_file_id: coverFileId,
  })
}
