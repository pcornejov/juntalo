package dto

type StartContributionRequest struct {
	FullName    string `json:"full_name" validate:"required,min=2,max=120"`
	Email       string `json:"email" validate:"omitempty,email"`
	Phone       string `json:"phone"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	IsAnonymous bool   `json:"is_anonymous"`
	Message     string `json:"message"`
}

type StartContributionResponse struct {
	ContributionID string `json:"contribution_id"`
	Payment        struct {
		Status      string `json:"status"`
		RedirectURL string `json:"redirect_url,omitempty"`
	} `json:"payment"`
}

type ContributionStatusResponse struct {
	Status string `json:"status"`
}

type WebhookPayload struct {
	ProviderRef string `json:"provider_ref"`
	Event       string `json:"event"`
}
