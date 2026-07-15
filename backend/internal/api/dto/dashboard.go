package dto

import "time"

type ParticipantResponse struct {
	ContributionID string    `json:"contribution_id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Amount         int64     `json:"amount"`
	IsAnonymous    bool      `json:"is_anonymous"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	Message        string    `json:"message,omitempty"`
	RaffleNumber   *int      `json:"raffle_number,omitempty"`
}

type ParticipantsResponse struct {
	Items   []ParticipantResponse `json:"items"`
	HasMore bool                  `json:"has_more"`
}
