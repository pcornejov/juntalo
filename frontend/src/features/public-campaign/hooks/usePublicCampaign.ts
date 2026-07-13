import { useQuery } from '@tanstack/react-query'
import { getPublicCampaign } from '../api'

export function usePublicCampaign(slug: string | undefined) {
  return useQuery({
    queryKey: ['public-campaign', slug],
    queryFn: () => getPublicCampaign(slug!),
    enabled: !!slug,
  })
}
