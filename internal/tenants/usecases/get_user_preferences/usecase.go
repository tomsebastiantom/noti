package getuserpreferences

import (
	"context"
	"fmt"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
	"github.com/google/uuid"
)

type GetUserPreferencesUseCase interface {
	Execute(ctx context.Context, req GetUserPreferencesRequest) (GetUserPreferencesResponse, error)
}

type getUserPreferencesUseCase struct {
	userPrefRepo repository.UserPreferenceRepository
	userRepo     repository.UserRepository
}

func NewGetUserPreferencesUseCase(
	userPrefRepo repository.UserPreferenceRepository,
	userRepo repository.UserRepository,
) GetUserPreferencesUseCase {
	return &getUserPreferencesUseCase{
		userPrefRepo: userPrefRepo,
		userRepo:     userRepo,
	}
}

func (uc *getUserPreferencesUseCase) Execute(ctx context.Context, req GetUserPreferencesRequest) (GetUserPreferencesResponse, error) {
	// Validate that the user exists
	_, err := uc.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return GetUserPreferencesResponse{}, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Get the user preferences
	userPref, err := uc.userPrefRepo.GetUserPreferenceByUserID(ctx, req.UserID)
	
	// If user preferences not found, create default preferences
	if err != nil {
		// Create default preferences
		defaultPref := domain.NewUserPreference(req.UserID, req.TenantID)
		defaultPref.ID = uuid.New().String()
		
		// Save the default preferences
		err = uc.userPrefRepo.CreateUserPreference(ctx, *defaultPref)
		if err != nil {
			return GetUserPreferencesResponse{}, fmt.Errorf("failed to create default preferences: %w", err)
		}
		
		// Return the default preferences
		return FromDomain(*defaultPref), nil
	}
	
	// Return existing preferences
	return FromDomain(userPref), nil
}
