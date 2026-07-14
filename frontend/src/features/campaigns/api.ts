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
  category: string
  title: string
  slug: string
  description: string
  cover_url?: string
  images: CampaignImage[]
  goal_amount?: number
  status: 'draft' | 'active' | 'paused' | 'finished' | 'suspended'
  starts_at?: string
  ends_at?: string
  publish_at?: string
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

export interface CampaignCategory {
  key: string
  label: string
  icon: string
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
  category?: string
  title: string
  description?: string
  goal_amount?: number
  publish_at?: string
}) {
  return apiClient.post<Campaign>('/campaigns', input)
}

export function cancelScheduledPublish(id: string) {
  return apiClient.post<Campaign>(`/campaigns/${id}/cancel-schedule`)
}

// updateCampaign siempre manda title/goal_amount/starts_at/ends_at aunque
// solo se edite la descripción: el backend sobreescribe esos campos tal
// cual llegan (a diferencia de cover_file_id/publish_at, que si mantienen
// el valor existente cuando viene null) — omitirlos los borraría.
export function updateCampaign(
  id: string,
  input: {
    title: string
    description: string
    category?: string
    goal_amount?: number
    starts_at?: string
    ends_at?: string
  },
) {
  return apiClient.patch<Campaign>(`/campaigns/${id}`, input)
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

export function cloneCampaign(id: string) {
  return apiClient.post<Campaign>(`/campaigns/${id}/clone`)
}

export function listCampaignTypes() {
  return apiClient.get<{ items: CampaignType[] }>('/meta/campaign-types')
}

export function listCampaignCategories() {
  return apiClient.get<{ items: CampaignCategory[] }>('/meta/campaign-categories')
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

export function listParticipants(
  campaignId: string,
  offset = 0,
  limit = PARTICIPANTS_PAGE_SIZE,
  search = '',
  status = '',
) {
  const params = new URLSearchParams({ limit: String(limit), offset: String(offset) })
  if (search) params.set('q', search)
  if (status) params.set('status', status)
  return apiClient.get<{ items: Participant[]; has_more: boolean }>(
    `/campaigns/${campaignId}/contributions?${params.toString()}`,
  )
}

export interface RefundResult {
  payment_id: string
  status: string
  amount: number
}

export function refundContribution(campaignId: string, contributionId: string, amount: number, reason?: string) {
  return apiClient.post<RefundResult>(`/campaigns/${campaignId}/contributions/${contributionId}/refund`, {
    amount,
    reason: reason ?? '',
  })
}
