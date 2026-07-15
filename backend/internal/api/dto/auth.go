package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2,max=120"`
	// Password solo lleva "required" a nivel de DTO — el largo mínimo lo
	// valida identity.ValidatePassword en el dominio (QA: un "min=8" acá
	// interceptaba antes y el código weak_password nunca llegaba a
	// devolverse a la API, pese a estar documentado en errors.go).
	Password string `json:"password" validate:"required"`
	// CaptchaToken no lleva "required": cuando TURNSTILE_SECRET_KEY está
	// vacía (dev/test) el verificador es un no-op que aprueba igual, sin
	// token — el enforcement real vive en RegisterService, no acá.
	CaptchaToken string `json:"captcha_token"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	EmailVerified bool   `json:"email_verified"`
	// IsAdmin: true si el email está en la lista blanca del backoffice
	// (ADMIN_EMAILS) — el frontend lo usa para mostrar/ocultar el link al
	// backoffice, pero el guardarraíl real vive en el backend
	// (middleware.RequireAdminUser), no acá.
	IsAdmin bool `json:"is_admin"`
}

type OrganizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Kind string `json:"kind"`
	// Datos de transferencia — vacíos hasta que el organizador los completa
	// desde su panel (Etapa post-MVP: liquidación manual).
	Rut                 string `json:"rut,omitempty"`
	PayoutBank          string `json:"payout_bank,omitempty"`
	PayoutAccountType   string `json:"payout_account_type,omitempty"`
	PayoutAccountNumber string `json:"payout_account_number,omitempty"`
	PayoutHolderName    string `json:"payout_holder_name,omitempty"`
}

type AuthResponse struct {
	User         UserResponse         `json:"user"`
	Organization OrganizationResponse `json:"organization"`
	AccessToken  string               `json:"access_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
	// ResetToken solo se llena cuando EXPOSE_RESET_LINKS=true — no hay envío
	// de email real todavía, así que este es el atajo explícito para poder
	// probar el flujo completo en este deploy de prueba (Config.ExposeResetLinks).
	// El frontend arma el link como /reset-password?token=<esto>.
	ResetToken string `json:"reset_token,omitempty"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}
