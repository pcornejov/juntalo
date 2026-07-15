import { useMutation } from '@tanstack/react-query'
import * as organizationsApi from '../api'

export function useUpdatePayoutInfo() {
  return useMutation({
    mutationFn: (input: organizationsApi.PayoutInfo) => organizationsApi.updatePayoutInfo(input),
  })
}
