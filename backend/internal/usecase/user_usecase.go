package usecase

import (
	"context"
	"fmt"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// UserUseCase implements interfaces.UserService.
type UserUseCase struct {
	users interfaces.UserRepository
}

// NewUser creates a new UserUseCase.
func NewUser(users interfaces.UserRepository) *UserUseCase {
	return &UserUseCase{users: users}
}

// GetProfile returns the authenticated user's profile.
func (uc *UserUseCase) GetProfile(ctx context.Context, userID string) (*dto.UserDTO, error) {
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return toUserDTO(user), nil
}

// UpdateProfile updates mutable profile fields.
func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserDTO, error) {
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Timezone != "" {
		user.Timezone = req.Timezone
	}
	if req.Currency != "" {
		user.Currency = req.Currency
	}
	if req.PreferredLanguage != "" {
		user.PreferredLanguage = req.PreferredLanguage
	}
	if req.ProfilePicture != "" {
		user.ProfilePicture = req.ProfilePicture
	}
	if err := uc.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update profile: save: %w", err)
	}
	return toUserDTO(user), nil
}

// DeleteAccount soft-deletes the user's account.
func (uc *UserUseCase) DeleteAccount(ctx context.Context, userID string) error {
	return uc.users.Delete(ctx, userID)
}

func toUserDTO(u *domain.User) *dto.UserDTO {
	return &dto.UserDTO{
		ID:                u.ID,
		Name:              u.Name,
		Email:             u.Email,
		ProfilePicture:    u.ProfilePicture,
		IsEmailVerified:   u.IsEmailVerified,
		Timezone:          u.Timezone,
		Currency:          u.Currency,
		PreferredLanguage: u.PreferredLanguage,
		CreatedAt:         u.CreatedAt,
	}
}
