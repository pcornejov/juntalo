import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as campaignsApi from '../api'

const campaignsKey = ['campaigns']
const campaignTypesKey = ['campaign-types']

export function useCampaigns() {
  return useQuery({
    queryKey: campaignsKey,
    queryFn: () => campaignsApi.listCampaigns().then((r) => r.items),
  })
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
  }
}

export function useParticipants(campaignId: string) {
  return useQuery({
    queryKey: ['participants', campaignId],
    queryFn: () => campaignsApi.listParticipants(campaignId).then((r) => r.items),
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
