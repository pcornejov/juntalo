// Package audit lists the stable action codes written to audit_logs
// (Etapa 5: auditoría enfocada en cambios de estado de campaña).
package audit

const (
	ActionCampaignCreated   = "campaign.created"
	ActionCampaignPublished = "campaign.published"
	ActionCampaignPaused    = "campaign.paused"
	ActionCampaignResumed   = "campaign.resumed"
	ActionCampaignFinished  = "campaign.finished"
	ActionCampaignDeleted   = "campaign.deleted"

	EntityTypeCampaign = "campaign"
)
