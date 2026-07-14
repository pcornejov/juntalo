import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import {
  EXPLORE_CAMPAIGNS_PAGE_SIZE,
  getFeaturedCampaign,
  getOrgProfile,
  getPublicCampaign,
  listPublicCampaigns,
  ORG_PROFILE_PAGE_SIZE,
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

// useOrgProfile pagina las campañas del organizador con el mismo patrón que
// useExploreCampaigns; nombre/slug/verificación se toman de la última
// página cargada (son constantes entre páginas de la misma organización).
export function useOrgProfile(slug: string | undefined) {
  const query = useInfiniteQuery({
    queryKey: ['org-profile', slug],
    queryFn: ({ pageParam }) => getOrgProfile(slug!, pageParam, ORG_PROFILE_PAGE_SIZE),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * ORG_PROFILE_PAGE_SIZE : undefined,
    enabled: !!slug,
  })
  const lastPage = query.data?.pages[query.data.pages.length - 1]
  return {
    ...query,
    name: lastPage?.name,
    orgSlug: lastPage?.slug,
    isVerified: lastPage?.is_verified ?? false,
    campaigns: query.data?.pages.flatMap((p) => p.campaigns) ?? [],
  }
}
