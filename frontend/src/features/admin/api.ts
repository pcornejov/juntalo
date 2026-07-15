import { apiClient } from '../../shared/api/client'

export interface AdminMetrics {
  total_users: number
  total_campaigns: number
  active_campaigns: number
  draft_campaigns: number
  finished_campaigns: number
  total_contributions: number
  raised_gross: number
  raised_net_approx: number
  total_commission: number
}

export interface AdminUser {
  id: string
  email: string
  full_name: string
  email_verified: boolean
  created_at: string
  organization_id: string
  organization_name: string
  organization_commission_rate: number
  campaign_count: number
}

export interface AdminCampaign {
  id: string
  title: string
  slug: string
  type_key: string
  status: string
  category: string
  created_at: string
  organization_name: string
  organizer_email: string
  raised_gross: number
  contributor_count: number
}

export interface AdminPayment {
  id: string
  status: string
  provider: string
  amount_gross: number
  amount_net: number
  commission_amount: number
  created_at: string
  confirmed_at?: string
  campaign_title: string
  campaign_slug: string
  contributor_name: string
}

export interface PendingPayout {
  organization_id: string
  organization_name: string
  rut: string
  payout_bank: string
  payout_account_type: string
  payout_account_number: string
  payout_holder_name: string
  eligible_net: number
  total_paid: number
  pending_amount: number
}

export const ADMIN_PAGE_SIZE = 20

export function getAdminMetrics() {
  return apiClient.get<AdminMetrics>('/admin/metrics')
}

export function listAdminUsers(offset = 0, limit = ADMIN_PAGE_SIZE) {
  return apiClient.get<{ items: AdminUser[]; has_more: boolean }>(
    `/admin/users?limit=${limit}&offset=${offset}`,
  )
}

export function listAdminCampaigns(offset = 0, limit = ADMIN_PAGE_SIZE) {
  return apiClient.get<{ items: AdminCampaign[]; has_more: boolean }>(
    `/admin/campaigns?limit=${limit}&offset=${offset}`,
  )
}

// deleteAdminCampaign es la herramienta de moderación del backoffice: a
// diferencia del borrado propio del organizador (solo borradores), un admin
// puede eliminar cualquier campaña sin importar su estado.
export function deleteAdminCampaign(id: string) {
  return apiClient.delete<void>(`/admin/campaigns/${id}`)
}

// updateOrgCommissionRate: herramienta del backoffice para ajustar la
// comisión de una organización sin tocar código. rate va como fracción
// (0.05 = 5%) — la conversión desde porcentaje vive en la UI.
export function updateOrgCommissionRate(orgId: string, rate: number) {
  return apiClient.patch<void>(`/admin/organizations/${orgId}/commission`, { rate })
}

export function listAdminPayments(offset = 0, limit = ADMIN_PAGE_SIZE) {
  return apiClient.get<{ items: AdminPayment[]; has_more: boolean }>(
    `/admin/payments?limit=${limit}&offset=${offset}`,
  )
}

export function listPendingPayouts(offset = 0, limit = ADMIN_PAGE_SIZE) {
  return apiClient.get<{ items: PendingPayout[]; has_more: boolean }>(
    `/admin/payouts/pending?limit=${limit}&offset=${offset}`,
  )
}

// createPayout deja constancia de una transferencia manual ya hecha por el
// operador — no dispara ningún movimiento de dinero real.
export function createPayout(orgId: string, amount: number, note?: string) {
  return apiClient.post<{ id: string }>(`/admin/organizations/${orgId}/payouts`, { amount, note })
}
