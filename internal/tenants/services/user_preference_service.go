package tenants

import (
	"context"
	"fmt"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
)

// UserPreferenceService handles checking user preferences for notifications
type UserPreferenceService struct {
	dbManager          *db.Manager
	logger             logger.Logger
	repositoryFactory  interface {
		GetUserPreferenceRepositoryForTenant(tenantID string) (repository.UserPreferenceRepository, error)
	}
}

// NewUserPreferenceService creates a new user preference service
func NewUserPreferenceService(
	dbManager *db.Manager,
	logger logger.Logger,
	repositoryFactory interface {
		GetUserPreferenceRepositoryForTenant(tenantID string) (repository.UserPreferenceRepository, error)
	},
) *UserPreferenceService {
	return &UserPreferenceService{
		dbManager:         dbManager,
		logger:            logger,
		repositoryFactory: repositoryFactory,
	}
}

// ShouldSendNotification checks if a notification should be sent based on user preferences
func (s *UserPreferenceService) ShouldSendNotification(
	ctx context.Context, 
	userID string, 
	tenantID string, 
	channel string, 
	category string,
) (bool, error) {
	if userID == "" {
		// No user ID, so we can't check preferences - default to sending
		return true, nil
	}

	// Get tenant-specific repository
	userPrefRepo, err := s.repositoryFactory.GetUserPreferenceRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get user preference repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.String("error",err.Error()),
			logger.Err(err))
		return true, fmt.Errorf("failed to get tenant database: %w", err)
	}

	// Get user preferences
	userPreference, err := userPrefRepo.GetUserPreferenceByUserID(ctx, userID)
	if err != nil {
		// If preferences don't exist or can't be retrieved, default to sending
		s.logger.InfoContext(ctx, "User preferences not found, defaulting to sending notification",
			logger.String("user_id", userID),
			logger.String("tenant_id", tenantID))
		return true, nil
	}

	// First check: is notifications enabled at all?
	if !userPreference.Enabled {
		s.logger.DebugContext(ctx, "Notifications disabled for user",
			logger.String("user_id", userID),
			logger.String("tenant_id", tenantID))
		return false, nil
	}

	// Check channel-level preference
	channelType := domain.ChannelType(channel)
	if channelEnabled, exists := userPreference.ChannelPrefs[channelType]; exists && !channelEnabled {
		s.logger.DebugContext(ctx, "Channel disabled for user",
			logger.String("user_id", userID),
			logger.String("channel", channel))
		return false, nil
	}

	// If there's a category, check category-level preferences
	if category != "" {
		if categoryPref, exists := userPreference.CategoryPrefs[category]; exists {
			// If category is disabled entirely
			if !categoryPref.Enabled {
				s.logger.DebugContext(ctx, "Category disabled for user",
					logger.String("user_id", userID),
					logger.String("category", category))
				return false, nil
			}

			// Check for channel-specific setting within this category
			if channelEnabled, exists := categoryPref.ChannelPrefs[channelType]; exists && !channelEnabled {
				s.logger.DebugContext(ctx, "Channel disabled for category for user",
					logger.String("user_id", userID),
					logger.String("channel", channel),
					logger.String("category", category))
				return false, nil
			}
		}
	}

	// All checks passed, notification can be sent
	return true, nil
}
