package providers

import (
	"context"

	providerDomain "getnoti.com/internal/providers/domain"
	tenantServices "getnoti.com/internal/tenants/services"
)

// UserPreferenceCheckerAdapter adapts the UserPreferenceService to the UserPreferenceChecker interface
// This follows the Adapter pattern to bridge between domains
type UserPreferenceCheckerAdapter struct {
	userPreferenceService *tenantServices.UserPreferenceService
}

// NewUserPreferenceCheckerAdapter creates a new adapter for user preferences
func NewUserPreferenceCheckerAdapter(userPrefService *tenantServices.UserPreferenceService) providerDomain.UserPreferenceChecker {
	return &UserPreferenceCheckerAdapter{
		userPreferenceService: userPrefService,
	}
}

// ShouldSendNotification implements the UserPreferenceChecker interface
func (a *UserPreferenceCheckerAdapter) ShouldSendNotification(ctx context.Context, userID, tenantID, channel, category string) (bool, error) {
	return a.userPreferenceService.ShouldSendNotification(ctx, userID, tenantID, channel, category)
}
