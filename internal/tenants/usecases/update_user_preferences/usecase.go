package updateuserpreferences

import (
	"context"
	"fmt"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
	"github.com/google/uuid"
)

type UpdateUserPreferencesUseCase interface {
	Execute(ctx context.Context, req UpdateUserPreferencesRequest) (UpdateUserPreferencesResponse, error)
}

type updateUserPreferencesUseCase struct {
	userPrefRepo repository.UserPreferenceRepository
	userRepo     repository.UserRepository
}

func NewUpdateUserPreferencesUseCase(
	userPrefRepo repository.UserPreferenceRepository,
	userRepo repository.UserRepository,
) UpdateUserPreferencesUseCase {
	return &updateUserPreferencesUseCase{
		userPrefRepo: userPrefRepo,
		userRepo:     userRepo,
	}
}

func (uc *updateUserPreferencesUseCase) Execute(ctx context.Context, req UpdateUserPreferencesRequest) (UpdateUserPreferencesResponse, error) {
	// Validate that the user exists
	_, err := uc.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return UpdateUserPreferencesResponse{Success: false}, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Try to get existing preferences
	existingPref, err := uc.userPrefRepo.GetUserPreferenceByUserID(ctx, req.UserID)
	var isNewPreference bool
	
	// If preferences don't exist yet, create new ones
	if err != nil {
		existingPref = *domain.NewUserPreference(req.UserID, req.TenantID)
		existingPref.ID = uuid.New().String()
		isNewPreference = true
	}
	
	// Update with requested changes
	existingPref.Enabled = req.Enabled
	
	// Update channel preferences if provided
	if req.ChannelPrefs != nil {
		existingPref.ChannelPrefs = req.ChannelPrefs
	}
	
	// Update category preferences if provided
	if req.CategoryPrefs != nil {
		existingPref.CategoryPrefs = req.CategoryPrefs
	}
	
	// Update digest settings if provided
	if req.DigestSettings != nil {
		existingPref.DigestSettings = *req.DigestSettings
	}
	
	// Validate the updated preferences
	if err := existingPref.Validate(); err != nil {
		return UpdateUserPreferencesResponse{Success: false}, fmt.Errorf("invalid preferences: %w", err)
	}
	
	// Save the preferences
	var saveErr error
	if isNewPreference {
		saveErr = uc.userPrefRepo.CreateUserPreference(ctx, existingPref)
	} else {
		saveErr = uc.userPrefRepo.UpdateUserPreference(ctx, existingPref)
	}
	
	if saveErr != nil {
		return UpdateUserPreferencesResponse{Success: false}, fmt.Errorf("failed to save preferences: %w", saveErr)
	}
	
	return UpdateUserPreferencesResponse{
		Success: true,
		Message: "User preferences updated successfully",
	}, nil
}
