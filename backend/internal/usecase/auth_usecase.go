// Package usecase contains the core business logic for authentication.
package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
	"github.com/priyanjul/ai-finance-tracker/pkg/auth"
	"github.com/priyanjul/ai-finance-tracker/pkg/email"
)

// AuthUseCase implements interfaces.AuthService.
type AuthUseCase struct {
	users       interfaces.UserRepository
	sessions    interfaces.SessionRepository
	emailSvc    *email.Service
	jwtSecret   string
	refreshSec  string
	jwtExpiry   time.Duration
	refreshExp  time.Duration
	appBaseURL  string
	googleOAuth *oauth2.Config
}

// NewAuth creates a new AuthUseCase with all dependencies injected.
func NewAuth(
	users interfaces.UserRepository,
	sessions interfaces.SessionRepository,
	emailSvc *email.Service,
	jwtSecret, refreshSec string,
	jwtExpiry, refreshExp time.Duration,
	appBaseURL string,
	googleClientID, googleClientSecret, googleRedirectURL string,
) *AuthUseCase {
	googleCfg := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &AuthUseCase{
		users: users, sessions: sessions, emailSvc: emailSvc,
		jwtSecret: jwtSecret, refreshSec: refreshSec,
		jwtExpiry: jwtExpiry, refreshExp: refreshExp,
		appBaseURL: appBaseURL,
		googleOAuth: googleCfg,
	}
}

// Signup creates a new user account, sends a verification email.
func (uc *AuthUseCase) Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error) {
	// Check if email is already registered
	existing, _ := uc.users.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("signup: hash password: %w", err)
	}

	verifyToken, _ := auth.GenerateSecureToken(32)

	user := &domain.User{
		ID:               uuid.NewString(),
		Email:            req.Email,
		Name:             req.Name,
		PasswordHash:     hash,
		EmailVerifyToken: verifyToken,
		IsEmailVerified:  false,
		Timezone:         "Asia/Kolkata",
		Currency:         "INR",
		PreferredLanguage: "en",
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("signup: create user: %w", err)
	}

	// Send verification email asynchronously (best-effort)
	link := fmt.Sprintf("%s/api/v1/auth/verify-email/%s", uc.appBaseURL, verifyToken)
	go func() {
		_ = uc.emailSvc.SendVerification(context.Background(), user.Email, link)
	}()

	return uc.issueTokens(ctx, user)
}

// Login verifies credentials and returns a token pair.
func (uc *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := uc.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return uc.issueTokens(ctx, user)
}

// RefreshToken validates a refresh token and issues a new token pair.
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	claims, err := auth.VerifyToken(refreshToken, uc.refreshSec)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	session, err := uc.sessions.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}
	if time.Now().After(session.RefreshExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	user, err := uc.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Invalidate old session
	_ = uc.sessions.DeleteByID(ctx, session.ID)

	return uc.issueTokens(ctx, user)
}

// VerifyEmail marks the user's email as verified.
func (uc *AuthUseCase) VerifyEmail(ctx context.Context, token string) error {
	// We need a GetByVerifyToken method; using a quick scan here
	// In production, add a repository method for this lookup
	return nil // TODO: implement verify token lookup
}

// ForgotPassword sends a password-reset link.
func (uc *AuthUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		// Don't leak whether the email exists
		return nil
	}

	token, _ := auth.GenerateSecureToken(32)
	expiry := time.Now().Add(1 * time.Hour)
	user.ResetToken = token
	user.ResetTokenExpiry = &expiry
	if err := uc.users.Update(ctx, user); err != nil {
		return err
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", uc.appBaseURL, token)
	go func() {
		_ = uc.emailSvc.SendPasswordReset(context.Background(), user.Email, link)
	}()
	return nil
}

// ResetPassword sets a new password using a reset token.
func (uc *AuthUseCase) ResetPassword(ctx context.Context, req dto.PasswordResetConfirm) error {
	// TODO: find user by reset token, validate expiry, hash and save new password
	return nil
}

// ChangePassword changes the authenticated user's password.
func (uc *AuthUseCase) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !auth.CheckPassword(req.OldPassword, user.PasswordHash) {
		return fmt.Errorf("incorrect current password")
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return uc.users.Update(ctx, user)
}

// Logout deletes all sessions for the user.
func (uc *AuthUseCase) Logout(ctx context.Context, userID string) error {
	return uc.sessions.DeleteByUserID(ctx, userID)
}

// GoogleOAuth handles sign-in via Google OAuth2 token exchange.
func (uc *AuthUseCase) GoogleOAuth(ctx context.Context, code string) (*dto.AuthResponse, error) {
	// 1. Exchange authorization code for tokens
	token, err := uc.googleOAuth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google oauth: token exchange: %w", err)
	}

	// 2. Fetch user info from Google
	client := uc.googleOAuth.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("google oauth: fetch userinfo: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var info struct {
		Sub     string `json:"sub"`   // Google user ID
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &info); err != nil || info.Email == "" {
		return nil, fmt.Errorf("google oauth: parse userinfo: %w", err)
	}

	// 3. Upsert user — find by email or GoogleID, create if new
	user, err := uc.users.GetByEmail(ctx, info.Email)
	if err != nil {
		// New user — create account
		user = &domain.User{
			ID:              uuid.NewString(),
			Email:           info.Email,
			Name:            info.Name,
			GoogleID:        info.Sub,
			ProfilePicture:  info.Picture,
			IsEmailVerified: true, // Google already verified the email
			Timezone:        "Asia/Kolkata",
			Currency:        "INR",
			PreferredLanguage: "en",
		}
		if err := uc.users.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("google oauth: create user: %w", err)
		}
	} else {
		// Existing user — update Google fields if not already set
		updated := false
		if user.GoogleID == "" {
			user.GoogleID = info.Sub
			updated = true
		}
		if user.ProfilePicture == "" && info.Picture != "" {
			user.ProfilePicture = info.Picture
			updated = true
		}
		if !user.IsEmailVerified {
			user.IsEmailVerified = true
			updated = true
		}
		if updated {
			_ = uc.users.Update(ctx, user)
		}
	}

	// 4. Issue JWT token pair
	return uc.issueTokens(ctx, user)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (uc *AuthUseCase) issueTokens(ctx context.Context, user *domain.User) (*dto.AuthResponse, error) {
	pair, err := auth.GenerateTokenPair(
		user.ID, user.Email,
		uc.jwtSecret, uc.refreshSec,
		uc.jwtExpiry, uc.refreshExp,
	)
	if err != nil {
		return nil, fmt.Errorf("issueTokens: %w", err)
	}

	session := &domain.Session{
		ID:               uuid.NewString(),
		UserID:           user.ID,
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		ExpiresAt:        pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}
	if err := uc.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("issueTokens: create session: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    int64(uc.jwtExpiry.Seconds()),
		TokenType:    "Bearer",
		User: dto.UserDTO{
			ID:                user.ID,
			Name:              user.Name,
			Email:             user.Email,
			ProfilePicture:    user.ProfilePicture,
			IsEmailVerified:   user.IsEmailVerified,
			Timezone:          user.Timezone,
			Currency:          user.Currency,
			PreferredLanguage: user.PreferredLanguage,
			CreatedAt:         user.CreatedAt,
		},
	}, nil
}
