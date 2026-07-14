package dto

import "time"

type AdminUserResponse struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	FullName         string    `json:"full_name"`
	EmailVerified    bool      `json:"email_verified"`
	CreatedAt        time.Time `json:"created_at"`
	OrganizationID   string    `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	CampaignCount    int64     `json:"campaign_count"`
}

type AdminUserListResponse struct {
	Items   []AdminUserResponse `json:"items"`
	HasMore bool                `json:"has_more"`
}

type AdminCampaignResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	TypeKey          string    `json:"type_key"`
	Status           string    `json:"status"`
	Category         string    `json:"category"`
	CreatedAt        time.Time `json:"created_at"`
	OrganizationName string    `json:"organization_name"`
	OrganizerEmail   string    `json:"organizer_email"`
	RaisedGross      int64     `json:"raised_gross"`
	ContributorCount int64     `json:"contributor_count"`
}

type AdminCampaignListResponse struct {
	Items   []AdminCampaignResponse `json:"items"`
	HasMore bool                    `json:"has_more"`
}

type AdminPaymentResponse struct {
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	Provider         string     `json:"provider"`
	AmountGross      int64      `json:"amount_gross"`
	AmountNet        int64      `json:"amount_net"`
	CommissionAmount int64      `json:"commission_amount"`
	CreatedAt        time.Time  `json:"created_at"`
	ConfirmedAt      *time.Time `json:"confirmed_at,omitempty"`
	CampaignTitle    string     `json:"campaign_title"`
	CampaignSlug     string     `json:"campaign_slug"`
	ContributorName  string     `json:"contributor_name"`
}

type AdminPaymentListResponse struct {
	Items   []AdminPaymentResponse `json:"items"`
	HasMore bool                   `json:"has_more"`
}

type AdminMetricsResponse struct {
	TotalUsers         int64 `json:"total_users"`
	TotalCampaigns     int64 `json:"total_campaigns"`
	ActiveCampaigns    int64 `json:"active_campaigns"`
	DraftCampaigns     int64 `json:"draft_campaigns"`
	FinishedCampaigns  int64 `json:"finished_campaigns"`
	TotalContributions int64 `json:"total_contributions"`
	RaisedGross        int64 `json:"raised_gross"`
	RaisedNetApprox    int64 `json:"raised_net_approx"`
	TotalCommission    int64 `json:"total_commission"`
}
