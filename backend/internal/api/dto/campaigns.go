package dto

import "time"

type CreateCampaignRequest struct {
	TypeKey     string     `json:"type_key" validate:"required"`
	Category    string     `json:"category"`
	Title       string     `json:"title" validate:"required,min=3,max=120"`
	Description string     `json:"description"`
	GoalAmount  *int64     `json:"goal_amount"`
	StartsAt    *time.Time `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	PublishAt   *time.Time `json:"publish_at"`
	// VideoURL es un link externo a YouTube o Vimeo — Juntalo no aloja video.
	VideoURL *string `json:"video_url"`
}

type UpdateCampaignRequest struct {
	Title          string     `json:"title" validate:"required,min=3,max=120"`
	Description    string     `json:"description"`
	Category       string     `json:"category"`
	GoalAmount     *int64     `json:"goal_amount"`
	StartsAt       *time.Time `json:"starts_at"`
	EndsAt         *time.Time `json:"ends_at"`
	CoverFileID    *string    `json:"cover_file_id"`
	PublishAt      *time.Time `json:"publish_at"`
	ClearPublishAt bool       `json:"clear_publish_at"`
	// VideoURL sigue el patrón "mantener si no viene" de CoverFileID;
	// ClearVideoURL es el único camino para quitar un video ya asociado.
	VideoURL      *string `json:"video_url"`
	ClearVideoURL bool    `json:"clear_video_url"`
}

type CampaignResponse struct {
	ID          string                  `json:"id"`
	TypeKey     string                  `json:"type_key"`
	Category    string                  `json:"category"`
	Title       string                  `json:"title"`
	Slug        string                  `json:"slug"`
	Description string                  `json:"description"`
	CoverURL    *string                 `json:"cover_url,omitempty"`
	Images      []CampaignImageResponse `json:"images"`
	GoalAmount  *int64                  `json:"goal_amount,omitempty"`
	Status      string                  `json:"status"`
	StartsAt    *time.Time              `json:"starts_at,omitempty"`
	EndsAt      *time.Time              `json:"ends_at,omitempty"`
	PublishAt   *time.Time              `json:"publish_at,omitempty"`
	PublicURL   string                  `json:"public_url"`
	Totals      TotalsDTO               `json:"totals"`
	CreatedAt   time.Time               `json:"created_at"`
	VideoURL    *string                 `json:"video_url,omitempty"`
}

type CampaignImageResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// ReorderImagesRequest lleva el nuevo orden completo de la galería; la
// primera imagen queda como portada.
type ReorderImagesRequest struct {
	ImageIDs []string `json:"image_ids" validate:"required,min=1"`
}

type TotalsDTO struct {
	RaisedGross      int64 `json:"raised_gross"`
	RaisedNetApprox  int64 `json:"raised_net_approx"`
	ContributorCount int64 `json:"contributor_count"`
}

type CampaignListResponse struct {
	Items   []CampaignResponse `json:"items"`
	HasMore bool               `json:"has_more"`
}

type PublicCampaignResponse struct {
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	CoverURL      *string   `json:"cover_url,omitempty"`
	Images        []string  `json:"images"`
	GoalAmount    *int64    `json:"goal_amount,omitempty"`
	Status        string    `json:"status"`
	Category      string    `json:"category"`
	Totals        TotalsDTO `json:"totals"`
	CTA           string    `json:"cta"`
	Unit          string    `json:"unit"`
	PublicURL     string    `json:"public_url"`
	OrganizerName string    `json:"organizer_name,omitempty"`
	IsVerified    bool      `json:"is_verified"`
	VideoURL      *string   `json:"video_url,omitempty"`
}

type PublicCampaignListResponse struct {
	Items   []PublicCampaignResponse `json:"items"`
	HasMore bool                     `json:"has_more"`
}

// OrgProfileResponse backs la página pública persistente del organizador
// (/org/:slug), inspirada en el link único de por vida de Ceneka.
type OrgProfileResponse struct {
	Name       string                   `json:"name"`
	Slug       string                   `json:"slug"`
	IsVerified bool                     `json:"is_verified"`
	Campaigns  []PublicCampaignResponse `json:"campaigns"`
	HasMore    bool                     `json:"has_more"`
}

type CampaignTypeResponse struct {
	Key                string `json:"key"`
	Name               string `json:"name"`
	CTA                string `json:"cta"`
	Unit               string `json:"unit"`
	RequiresGoalAmount bool   `json:"requires_goal_amount"`
	AllowsFreeAmount   bool   `json:"allows_free_amount"`
}

type CampaignCategoryResponse struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}
