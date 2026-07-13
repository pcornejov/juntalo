package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2,max=120"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type OrganizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type AuthResponse struct {
	User         UserResponse         `json:"user"`
	Organization OrganizationResponse `json:"organization"`
	AccessToken  string               `json:"access_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}
