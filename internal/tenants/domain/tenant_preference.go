package domain

import (
	"errors"
	"time"
)

// TenantPreference represents default notification preferences for a tenant
type TenantPreference struct {
	ID             string                        `json:"id"`
	TenantID       string                        `json:"tenantId"`
	Enabled        bool                          `json:"enabled"`          // Master toggle for all notifications
	ChannelPrefs   map[ChannelType]bool          `json:"channelPrefs"`     // Enable/disable specific channels
	CategoryPrefs  map[string]CategoryPreference `json:"categoryPrefs"`    // Preferences by notification category
	DigestSettings DigestSettings                `json:"digestSettings"`   // Default digest configuration
	CreatedAt      time.Time                     `json:"createdAt"`
	UpdatedAt      time.Time                     `json:"updatedAt"`
}

// NewTenantPreference creates a new TenantPreference with default values
func NewTenantPreference(tenantID string) *TenantPreference {
	return &TenantPreference{
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

// Validate checks if the tenant preference is valid
func (tp *TenantPreference) Validate() error {
	if tp.TenantID == "" {
		return errors.New("tenant ID cannot be empty")
	}
	
	return nil
}
