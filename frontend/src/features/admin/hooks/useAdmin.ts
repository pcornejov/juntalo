import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import * as adminApi from '../api'

export function useAdminMetrics() {
  return useQuery({
    queryKey: ['admin', 'metrics'],
    queryFn: adminApi.getAdminMetrics,
  })
}

export function useAdminUsers() {
  const query = useInfiniteQuery({
    queryKey: ['admin', 'users'],
    queryFn: ({ pageParam }) => adminApi.listAdminUsers(pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * adminApi.ADMIN_PAGE_SIZE : undefined,
  })
  return { ...query, items: query.data?.pages.flatMap((p) => p.items) ?? [] }
}

export function useAdminCampaigns() {
  const query = useInfiniteQuery({
    queryKey: ['admin', 'campaigns'],
    queryFn: ({ pageParam }) => adminApi.listAdminCampaigns(pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * adminApi.ADMIN_PAGE_SIZE : undefined,
  })
  return { ...query, items: query.data?.pages.flatMap((p) => p.items) ?? [] }
}

export function useAdminPayments() {
  const query = useInfiniteQuery({
    queryKey: ['admin', 'payments'],
    queryFn: ({ pageParam }) => adminApi.listAdminPayments(pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * adminApi.ADMIN_PAGE_SIZE : undefined,
  })
  return { ...query, items: query.data?.pages.flatMap((p) => p.items) ?? [] }
}
