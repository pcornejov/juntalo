import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import {
  EXPLORE_CAMPAIGNS_PAGE_SIZE,
  getFeaturedCampaign,
  getPublicCampaign,
  listPublicCampaigns,
} from '../api'

export function usePublicCampaign(slug: string | undefined) {
  return useQuery({
    queryKey: ['public-campaign', slug],
    queryFn: () => getPublicCampaign(slug!),
    enabled: !!slug,
  })
}

export function useExploreCampaigns(search: string, category: string) {
  const query = useInfiniteQuery({
    queryKey: ['explore-campaigns', search, category],
    queryFn: ({ pageParam }) =>
      listPublicCampaigns(pageParam, EXPLORE_CAMPAIGNS_PAGE_SIZE, search, category),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * EXPLORE_CAMPAIGNS_PAGE_SIZE : undefined,
  })
  return {
    ...query,
    campaigns: query.data?.pages.flatMap((p) => p.items) ?? [],
  }
}

// La campaña "más caliente" solo tiene sentido cuando no hay filtro activo
// (búsqueda/categoría) — es una recomendación editorial de toda la
// plataforma, no del subconjunto filtrado.
export function useFeaturedCampaign() {
  return useQuery({
    queryKey: ['featured-campaign'],
    queryFn: getFeaturedCampaign,
  })
}
