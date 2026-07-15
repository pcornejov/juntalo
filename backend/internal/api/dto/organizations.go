package dto

type UpdatePayoutInfoRequest struct {
	Rut                 string `json:"rut" validate:"required"`
	PayoutBank          string `json:"payout_bank" validate:"required"`
	PayoutAccountType   string `json:"payout_account_type" validate:"required"`
	PayoutAccountNumber string `json:"payout_account_number" validate:"required"`
	PayoutHolderName    string `json:"payout_holder_name" validate:"required"`
}

type PayoutInfoResponse struct {
	Rut                 string `json:"rut"`
	PayoutBank          string `json:"payout_bank"`
	PayoutAccountType   string `json:"payout_account_type"`
	PayoutAccountNumber string `json:"payout_account_number"`
	PayoutHolderName    string `json:"payout_holder_name"`
}
