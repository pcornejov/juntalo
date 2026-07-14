import { apiClient } from '../../shared/api/client'

export interface Totals {
  raised_gross: number
  raised_net_approx: number
  contributor_count: number
}

export interface CampaignImage {
  id: string
  url: string
}

export interface Campaign {
  id: string
  type_key: string
  title: string
  slug: string
  description: string
  cover_url?: string
  images: CampaignImage[]
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

export const CAMPAIGNS_PAGE_SIZE = 20

export function listCampaigns(offset = 0, limit = CAMPAIGNS_PAGE_SIZE) {
  return apiClient.get<{ items: Campaign[]; has_more: boolean }>(
    `/campaigns?limit=${limit}&offset=${offset}`,
  )
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

export function addCampaignImage(campaignId: string, file: File) {
  const form = new FormData()
  form.append('file', file)
  return apiClient.upload<{ id: string; url: string }>(`/campaigns/${campaignId}/images`, form)
}

export function deleteCampaignImage(campaignId: string, imageId: string) {
  return apiClient.delete<void>(`/campaigns/${campaignId}/images/${imageId}`)
}

export function reorderCampaignImages(campaignId: string, imageIds: string[]) {
  return apiClient.patch<Campaign>(`/campaigns/${campaignId}/images/reorder`, { image_ids: imageIds })
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

export const PARTICIPANTS_PAGE_SIZE = 20

export function listParticipants(campaignId: string, offset = 0, limit = PARTICIPANTS_PAGE_SIZE) {
  return apiClient.get<{ items: Participant[]; has_more: boolean }>(
    `/campaigns/${campaignId}/contributions?limit=${limit}&offset=${offset}`,
  )
}
