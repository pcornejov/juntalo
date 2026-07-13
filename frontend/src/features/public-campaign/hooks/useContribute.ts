import { useMemo } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as api from '../api'

// La Idempotency-Key se genera una vez al montar el formulario (Etapa 4 §1):
// reintentar el mismo submit (doble-tap, reconexión) nunca duplica el cargo.
export function useContribute(slug: string) {
  const idempotencyKey = useMemo(() => crypto.randomUUID(), [])
  const qc = useQueryClient()

  const mutation = useMutation({
    mutationFn: (input: api.StartContributionInput) => api.startContribution(slug, idempotencyKey, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['public-campaign', slug] }),
  })

  return { idempotencyKey, ...mutation }
}

export function useContributionStatus(contributionId: string | undefined, initialStatus: string | undefined) {
  return useQuery({
    queryKey: ['contribution-status', contributionId],
    queryFn: () => api.getContributionStatus(contributionId!),
    enabled: !!contributionId,
    initialData: initialStatus ? { status: initialStatus } : undefined,
    // Deja de pollear apenas el pago sale de pending (confirmado o fallido).
    refetchInterval: (query) => (query.state.data?.status === 'pending' ? 1000 : false),
  })
}
