// Package handler contains Gin HTTP handlers for all API endpoints.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
	"github.com/priyanjul/ai-finance-tracker/internal/middleware"
)

// ─── Auth Handler ─────────────────────────────────────────────────────────────

// AuthHandler exposes authentication endpoints.
type AuthHandler struct {
	auth interfaces.AuthService
}

// NewAuth creates a new AuthHandler.
func NewAuth(auth interfaces.AuthService) *AuthHandler { return &AuthHandler{auth: auth} }

// Signup godoc
// @Summary      Register new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.SignupRequest true "Signup payload"
// @Success      201  {object} dto.AuthResponse
// @Failure      400  {object} dto.ErrorResponse
// @Router       /auth/signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}
	resp, err := h.auth.Signup(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "SIGNUP_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary      Authenticate user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginRequest true "Login payload"
// @Success      200  {object} dto.AuthResponse
// @Failure      401  {object} dto.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}
	resp, err := h.auth.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: "INVALID_CREDENTIALS", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RefreshTokenRequest true "Refresh token"
// @Success      200  {object} dto.AuthResponse
// @Failure      401  {object} dto.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}
	resp, err := h.auth.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: "INVALID_REFRESH_TOKEN", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// VerifyEmail godoc
// @Summary      Verify email address
// @Tags         auth
// @Param        token path string true "Verification token"
// @Success      200
// @Failure      400  {object} dto.ErrorResponse
// @Router       /auth/verify-email/{token} [get]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	if err := h.auth.VerifyEmail(c.Request.Context(), c.Param("token")); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "VERIFY_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "email verified"})
}

// ForgotPassword godoc
// @Summary      Request password reset
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.PasswordResetRequest true "Email"
// @Success      200
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}
	_ = h.auth.ForgotPassword(c.Request.Context(), req.Email)
	// Always 200 – don't reveal whether email exists
	c.JSON(http.StatusOK, gin.H{"message": "if your email is registered, a reset link has been sent"})
}

// ResetPassword godoc
// @Summary      Reset password with token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.PasswordResetConfirm true "Reset payload"
// @Success      200
// @Failure      400  {object} dto.ErrorResponse
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.PasswordResetConfirm
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}
	if err := h.auth.ResetPassword(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "RESET_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

// ChangePassword godoc
// @Summary      Change authenticated user's password
// @Security     BearerAuth
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.ChangePasswordRequest true "Change password payload"
// @Success      200
// @Failure      400  {object} dto.ErrorResponse
// @Router       /users/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}
	if err := h.auth.ChangePassword(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "CHANGE_PASSWORD_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// Logout godoc
// @Summary      Logout (invalidate all sessions)
// @Security     BearerAuth
// @Tags         auth
// @Success      200
// @Router       /users/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	_ = h.auth.Logout(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// GoogleOAuthCallback handles the OAuth2 redirect from Google.
func (h *AuthHandler) GoogleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "MISSING_CODE", Message: "authorization code required"})
		return
	}
	resp, err := h.auth.GoogleOAuth(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "OAUTH_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ─── User Profile Handler ─────────────────────────────────────────────────────

// UserHandler exposes user profile endpoints.
type UserHandler struct {
	svc interfaces.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc interfaces.UserService) *UserHandler { return &UserHandler{svc: svc} }

// GetProfile returns the authenticated user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	profile, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates mutable profile fields.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}
	profile, err := h.svc.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "UPDATE_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}
