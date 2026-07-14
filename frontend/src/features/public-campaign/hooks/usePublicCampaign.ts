import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import { EXPLORE_CAMPAIGNS_PAGE_SIZE, getPublicCampaign, listPublicCampaigns } from '../api'

export function usePublicCampaign(slug: string | undefined) {
  return useQuery({
    queryKey: ['public-campaign', slug],
    queryFn: () => getPublicCampaign(slug!),
    enabled: !!slug,
  })
}

export function useExploreCampaigns(search: string) {
  const query = useInfiniteQuery({
    queryKey: ['explore-campaigns', search],
    queryFn: ({ pageParam }) => listPublicCampaigns(pageParam, EXPLORE_CAMPAIGNS_PAGE_SIZE, search),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * EXPLORE_CAMPAIGNS_PAGE_SIZE : undefined,
  })
  return {
    ...query,
    campaigns: query.data?.pages.flatMap((p) => p.items) ?? [],
  }
}
