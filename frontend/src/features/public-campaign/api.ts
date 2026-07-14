import { apiClient } from '../../shared/api/client'
import type { Totals } from '../campaigns/api'

export interface PublicCampaign {
  slug: string
  title: string
  description: string
  cover_url?: string
  images: string[]
  goal_amount?: number
  status: string
  category: string
  totals: Totals
  cta: string
  unit: string
  public_url: string
  created_at?: string
  // Campos aditivos y opcionales: el backend actual no los expone todavía.
  // Mientras no vengan en la respuesta, la UI oculta el bloque
  // correspondiente en vez de inventar datos — ver PublicCampaignPage.
  organizer_name?: string
  organizer_campaign_count?: number
  is_verified?: boolean
  location?: string
  video_url?: string
}

export function getPublicCampaign(slug: string) {
  return apiClient.get<PublicCampaign>(`/public/campaigns/${slug}`)
}

export const EXPLORE_CAMPAIGNS_PAGE_SIZE = 12

// listPublicCampaigns backs la sección "Explorar campañas": cualquier
// visitante puede navegar campañas activas de cualquier organizador, no
// solo entrar por un link directo (Etapa 4).
export function listPublicCampaigns(
  offset = 0,
  limit = EXPLORE_CAMPAIGNS_PAGE_SIZE,
  search = '',
  category = '',
) {
  const params = new URLSearchParams({ limit: String(limit), offset: String(offset) })
  if (search) params.set('q', search)
  if (category) params.set('category', category)
  return apiClient.get<{ items: PublicCampaign[]; has_more: boolean }>(
    `/public/campaigns?${params.toString()}`,
  )
}

// getFeaturedCampaign devuelve undefined si aún no hay campaña "más
// caliente" (backend responde 204 No Content, que apiClient.get resuelve a
// undefined) — inspirado en Vaki, se calcula por el aporte confirmado más
// reciente, no un campo manual del organizador.
export function getFeaturedCampaign() {
  return apiClient.get<PublicCampaign | undefined>('/public/campaigns/featured')
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
