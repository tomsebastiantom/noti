package domain

import (
	"errors"
	"time"
)

// ChannelType represents a notification delivery channel
type ChannelType string

// Different notification channels
const (
	ChannelTypeEmail     ChannelType = "email"
	ChannelTypeSMS       ChannelType = "sms"
	ChannelTypePush      ChannelType = "push"
	ChannelTypeWebPush   ChannelType = "web-push"
	ChannelTypeInApp     ChannelType = "in-app"
)

// DigestType represents how notification digests should be delivered
type DigestType string

const (
	DigestTypeNone     DigestType = "none"     // No digest, send immediately
	DigestTypeDaily    DigestType = "daily"    // Daily digest
	DigestTypeWeekly   DigestType = "weekly"   // Weekly digest
	DigestTypeInterval DigestType = "interval" // Custom interval-based digest
)

// UserPreference represents a user's notification preferences
type UserPreference struct {
	ID             string                        `json:"id"`
	UserID         string                        `json:"userId"`
	TenantID       string                        `json:"tenantId"`
	Enabled        bool                          `json:"enabled"`          // Master toggle for all notifications
	ChannelPrefs   map[ChannelType]bool          `json:"channelPrefs"`     // Enable/disable specific channels
	CategoryPrefs  map[string]CategoryPreference `json:"categoryPrefs"`    // Preferences by notification category
	DigestSettings DigestSettings                `json:"digestSettings"`   // Digest configuration
	CreatedAt      time.Time                     `json:"createdAt"`
	UpdatedAt      time.Time                     `json:"updatedAt"`
}

// CategoryPreference represents preferences for a specific notification category
type CategoryPreference struct {
	Enabled       bool                     `json:"enabled"`
	ChannelPrefs  map[ChannelType]bool     `json:"channelPrefs"` // Override channel preferences for this category
	DigestEnabled bool                     `json:"digestEnabled"`
	DigestType    DigestType               `json:"digestType"`
}

// DigestSettings represents a user's digest configuration
type DigestSettings struct {
	Enabled            bool       `json:"enabled"`
	Type               DigestType `json:"type"`
	IntervalMinutes    int        `json:"intervalMinutes"` // Used when Type is DigestTypeInterval
	DeliveryHour       int        `json:"deliveryHour"`    // Hour of the day (0-23) for delivering digests
	PreferredDayOfWeek int        `json:"preferredDayOfWeek"` // Day of the week (0=Sun, 6=Sat) for weekly digests
	PreferredChannel   ChannelType `json:"preferredChannel"`
}

// NewUserPreference creates a new UserPreference with default values
func NewUserPreference(userID, tenantID string) *UserPreference {
	return &UserPreference{
		UserID:   userID,
		TenantID: tenantID,
		Enabled:  true, // Enabled by default
		ChannelPrefs: map[ChannelType]bool{
			ChannelTypeEmail:   true,
			ChannelTypeSMS:     true,
			ChannelTypePush:    true,
			ChannelTypeWebPush: true,
			ChannelTypeInApp:   true,
		},
		CategoryPrefs: map[string]CategoryPreference{},
		DigestSettings: DigestSettings{
			Enabled:            false,
			Type:               DigestTypeNone,
			IntervalMinutes:    60,
			DeliveryHour:       9, // Default to 9 AM
			PreferredDayOfWeek: 1, // Default to Monday
			PreferredChannel:   ChannelTypeEmail,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Validate checks if the user preference is valid
func (up *UserPreference) Validate() error {
	if up.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	
	if up.TenantID == "" {
		return errors.New("tenant ID cannot be empty")
	}
	
	// Validate digest settings
	if up.DigestSettings.Enabled {
		switch up.DigestSettings.Type {
		case DigestTypeInterval:
			if up.DigestSettings.IntervalMinutes <= 0 {
				return errors.New("digest interval must be greater than 0")
			}
		case DigestTypeDaily:
			if up.DigestSettings.DeliveryHour < 0 || up.DigestSettings.DeliveryHour > 23 {
				return errors.New("delivery hour must be between 0 and 23")
			}
		case DigestTypeWeekly:
			if up.DigestSettings.PreferredDayOfWeek < 0 || up.DigestSettings.PreferredDayOfWeek > 6 {
				return errors.New("preferred day of week must be between 0 and 6")
			}
			if up.DigestSettings.DeliveryHour < 0 || up.DigestSettings.DeliveryHour > 23 {
				return errors.New("delivery hour must be between 0 and 23")
			}
		}
	}
	
	return nil
}
