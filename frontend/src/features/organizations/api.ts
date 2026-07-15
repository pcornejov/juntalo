import { apiClient } from '../../shared/api/client'

export interface PayoutInfo {
  rut: string
  payout_bank: string
  payout_account_type: string
  payout_account_number: string
  payout_holder_name: string
}

export function updatePayoutInfo(input: PayoutInfo) {
  return apiClient.patch<PayoutInfo>('/organizations/me/payout', input)
}
