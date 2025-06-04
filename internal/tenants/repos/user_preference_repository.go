package repository

import (
	"context"

	"getnoti.com/internal/tenants/domain"
)

type UserPreferenceRepository interface {
	// Create a new user preference
	CreateUserPreference(ctx context.Context, preference domain.UserPreference) error
	
	// Get user preference by user ID
	GetUserPreferenceByUserID(ctx context.Context, userID string) (domain.UserPreference, error)
	
	// Update an existing user preference
	UpdateUserPreference(ctx context.Context, preference domain.UserPreference) error
	
	// Get all user preferences for a specific category
	GetUserPreferencesByCategory(ctx context.Context, category string) ([]domain.UserPreference, error)
	
	// Get all users with digest settings matching certain criteria
	GetUsersForDigest(ctx context.Context, digestType domain.DigestType, dayOfWeek, hour int) ([]domain.UserPreference, error)
}
