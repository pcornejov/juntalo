import { useInfiniteQuery, useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as campaignsApi from '../api'

const campaignsKey = ['campaigns']
const campaignTypesKey = ['campaign-types']

// "Cargar más" en vez de números de página: a esta escala (campañas de un
// solo organizador) no vale la pena una UI de paginación numerada.
export function useCampaigns() {
  const query = useInfiniteQuery({
    queryKey: campaignsKey,
    queryFn: ({ pageParam }) => campaignsApi.listCampaigns(pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * campaignsApi.CAMPAIGNS_PAGE_SIZE : undefined,
  })
  return {
    ...query,
    campaigns: query.data?.pages.flatMap((p) => p.items) ?? [],
  }
}

export function useCampaign(id: string | undefined) {
  return useQuery({
    queryKey: [...campaignsKey, id],
    queryFn: () => campaignsApi.getCampaign(id!),
    enabled: !!id,
  })
}

export function useCampaignTypes() {
  return useQuery({
    queryKey: campaignTypesKey,
    queryFn: () => campaignsApi.listCampaignTypes().then((r) => r.items),
    staleTime: Infinity,
  })
}

export function useCreateCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: campaignsApi.createCampaign,
    onSuccess: () => qc.invalidateQueries({ queryKey: campaignsKey }),
  })
}

export function usePublishCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: campaignsApi.publishCampaign,
    onSuccess: () => qc.invalidateQueries({ queryKey: campaignsKey }),
  })
}

export function useCancelScheduledPublish() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: campaignsApi.cancelScheduledPublish,
    onSuccess: () => qc.invalidateQueries({ queryKey: campaignsKey }),
  })
}

export function useCloneCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: campaignsApi.cloneCampaign,
    onSuccess: () => qc.invalidateQueries({ queryKey: campaignsKey }),
  })
}

export function useCampaignGallery(campaignId: string) {
  const qc = useQueryClient()
  const invalidate = () => qc.invalidateQueries({ queryKey: [...campaignsKey, campaignId] })
  return {
    addImage: useMutation({
      mutationFn: (file: File) => campaignsApi.addCampaignImage(campaignId, file),
      onSuccess: invalidate,
    }),
    deleteImage: useMutation({
      mutationFn: (imageId: string) => campaignsApi.deleteCampaignImage(campaignId, imageId),
      onSuccess: invalidate,
    }),
    reorder: useMutation({
      mutationFn: (imageIds: string[]) => campaignsApi.reorderCampaignImages(campaignId, imageIds),
      onSuccess: invalidate,
    }),
  }
}

export function useParticipants(campaignId: string) {
  const query = useInfiniteQuery({
    queryKey: ['participants', campaignId],
    queryFn: ({ pageParam }) => campaignsApi.listParticipants(campaignId, pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.has_more ? allPages.length * campaignsApi.PARTICIPANTS_PAGE_SIZE : undefined,
  })
  return {
    ...query,
    participants: query.data?.pages.flatMap((p) => p.items) ?? [],
  }
}

export function useRefundContribution(campaignId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ contributionId, amount, reason }: { contributionId: string; amount: number; reason?: string }) =>
      campaignsApi.refundContribution(campaignId, contributionId, amount, reason),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['participants', campaignId] })
      qc.invalidateQueries({ queryKey: [...campaignsKey, campaignId] })
    },
  })
}

export function useCampaignTransitions() {
  const qc = useQueryClient()
  const invalidate = () => qc.invalidateQueries({ queryKey: campaignsKey })
  return {
    pause: useMutation({ mutationFn: campaignsApi.pauseCampaign, onSuccess: invalidate }),
    resume: useMutation({ mutationFn: campaignsApi.resumeCampaign, onSuccess: invalidate }),
    finish: useMutation({ mutationFn: campaignsApi.finishCampaign, onSuccess: invalidate }),
  }
}
