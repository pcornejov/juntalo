package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/api/middleware"
	"github.com/pcornejov/juntalo/backend/internal/app"
	authuc "github.com/pcornejov/juntalo/backend/internal/app/auth"
	"github.com/pcornejov/juntalo/backend/internal/domain/identity"
)

const (
	accessTokenTTL   = 15 * time.Minute
	refreshCookieKey = "jt_refresh"
)

type AuthHandler struct {
	register         *authuc.RegisterService
	login            *authuc.LoginService
	refresh          *authuc.RefreshService
	forgotPassword   *authuc.ForgotPasswordService
	resetPassword    *authuc.ResetPasswordService
	users            app.UserRepository
	orgs             app.OrganizationRepository
	signer           app.TokenSigner
	isProd           bool
	exposeResetLinks bool
}

func NewAuthHandler(
	register *authuc.RegisterService,
	login *authuc.LoginService,
	refresh *authuc.RefreshService,
	forgotPassword *authuc.ForgotPasswordService,
	resetPassword *authuc.ResetPasswordService,
	users app.UserRepository,
	orgs app.OrganizationRepository,
	signer app.TokenSigner,
	isProd bool,
	exposeResetLinks bool,
) *AuthHandler {
	return &AuthHandler{
		register:         register,
		login:            login,
		refresh:          refresh,
		forgotPassword:   forgotPassword,
		resetPassword:    resetPassword,
		users:            users,
		orgs:             orgs,
		signer:           signer,
		isProd:           isProd,
		exposeResetLinks: exposeResetLinks,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	user, org, err := h.register.Register(c.Context(), req.Email, req.FullName, req.Password)
	if err != nil {
		return dto.WriteError(c, err)
	}

	return h.issueSession(c, user, org, fiber.StatusCreated)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	user, org, err := h.login.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return dto.WriteError(c, err)
	}

	return h.issueSession(c, user, org, fiber.StatusOK)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	rawToken := c.Cookies(refreshCookieKey)
	if rawToken == "" {
		return dto.WriteError(c, sessionExpired())
	}

	userID, newRaw, err := h.refresh.Rotate(c.Context(), rawToken)
	if err != nil {
		return dto.WriteError(c, err)
	}

	accessToken, err := h.signer.Sign(userID, accessTokenTTL)
	if err != nil {
		return dto.WriteError(c, err)
	}

	h.setRefreshCookie(c, newRaw)
	return c.JSON(dto.RefreshResponse{AccessToken: accessToken})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	rawToken := c.Cookies(refreshCookieKey)
	if rawToken != "" {
		_ = h.refresh.Revoke(c.Context(), rawToken)
	}
	h.clearRefreshCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	token, err := h.forgotPassword.RequestReset(c.Context(), req.Email)
	if err != nil {
		return dto.WriteError(c, err)
	}

	// Mismo mensaje exista o no el email — evita que alguien pueda usar este
	// endpoint para averiguar qué correos están registrados.
	resp := dto.ForgotPasswordResponse{
		Message: "Si el email existe, te enviaremos instrucciones para restablecer tu contraseña.",
	}
	if h.exposeResetLinks && token != "" {
		resp.ResetToken = token
	}
	return c.JSON(resp)
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return dto.WriteError(c, err)
	}
	if err := dto.Validate(req); err != nil {
		return dto.WriteError(c, err)
	}

	if err := h.resetPassword.Reset(c.Context(), req.Token, req.NewPassword); err != nil {
		return dto.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := middleware.UserID(c)

	user, found, err := h.users.GetByID(c.Context(), userID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	if !found {
		return dto.WriteError(c, sessionExpired())
	}

	org, err := h.orgs.GetPersonalByUserID(c.Context(), userID)
	if err != nil {
		return dto.WriteError(c, err)
	}

	return c.JSON(fiber.Map{
		"user":         toUserResponse(user),
		"organization": toOrganizationResponse(org),
	})
}

func (h *AuthHandler) issueSession(c *fiber.Ctx, user identity.User, org identity.Organization, status int) error {
	accessToken, err := h.signer.Sign(user.ID, accessTokenTTL)
	if err != nil {
		return dto.WriteError(c, err)
	}

	rawRefresh, err := h.refresh.IssueRefreshToken(c.Context(), user.ID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	h.setRefreshCookie(c, rawRefresh)

	return c.Status(status).JSON(dto.AuthResponse{
		User:         toUserResponse(user),
		Organization: toOrganizationResponse(org),
		AccessToken:  accessToken,
	})
}

// cookieSameSite returns None in production: frontend y backend viven en
// dominios distintos (Render, Vercel, etc.) salvo que Caddy los unifique como
// en el diseño de VPS (Etapa 2). None requiere Secure=true siempre.
func (h *AuthHandler) cookieSameSite() string {
	if h.isProd {
		return fiber.CookieSameSiteNoneMode
	}
	return fiber.CookieSameSiteLaxMode
}

func (h *AuthHandler) setRefreshCookie(c *fiber.Ctx, rawToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieKey,
		Value:    rawToken,
		HTTPOnly: true,
		Secure:   h.isProd,
		SameSite: h.cookieSameSite(),
		Expires:  time.Now().Add(authuc.RefreshTokenTTL),
		Path:     "/api/v1/auth",
	})
}

func (h *AuthHandler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieKey,
		Value:    "",
		HTTPOnly: true,
		Secure:   h.isProd,
		SameSite: h.cookieSameSite(),
		Expires:  time.Now().Add(-time.Hour),
		Path:     "/api/v1/auth",
	})
}

func toUserResponse(u identity.User) dto.UserResponse {
	return dto.UserResponse{ID: u.ID.String(), Email: u.Email, FullName: u.FullName}
}

func toOrganizationResponse(o identity.Organization) dto.OrganizationResponse {
	return dto.OrganizationResponse{ID: o.ID.String(), Name: o.Name, Kind: o.Kind}
}

func sessionExpired() error {
	return authuc.ErrSessionExpired
}
