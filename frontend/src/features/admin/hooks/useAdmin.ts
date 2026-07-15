import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
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

export function useDeleteAdminCampaign() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => adminApi.deleteAdminCampaign(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'campaigns'] })
      queryClient.invalidateQueries({ queryKey: ['admin', 'metrics'] })
    },
  })
}

export function useUpdateOrgCommissionRate() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ orgId, rate }: { orgId: string; rate: number }) =>
      adminApi.updateOrgCommissionRate(orgId, rate),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
    },
  })
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

export function usePendingPayouts() {
  const query = useInfiniteQuery({
    queryKey: ['admin', 'payouts', 'pending'],
    queryFn: ({ pageParam }) => adminApi.listPendingPayouts(pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * adminApi.ADMIN_PAGE_SIZE : undefined,
  })
  return { ...query, items: query.data?.pages.flatMap((p) => p.items) ?? [] }
}

export function useCreatePayout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ orgId, amount, note }: { orgId: string; amount: number; note?: string }) =>
      adminApi.createPayout(orgId, amount, note),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'payouts', 'pending'] })
    },
  })
}
