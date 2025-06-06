package events

import (
	"getnoti.com/internal/shared/events"
)

// Tenant Event Types
const (
	TenantCreatedEventType                  = "tenant.created"
	TenantUpdatedEventType                  = "tenant.updated"
	TenantDeletedEventType                  = "tenant.deleted"
	TenantConfigurationUpdatedEventType     = "tenant.configuration.updated"
	UserCreatedEventType                    = "user.created"
	UserUpdatedEventType                    = "user.updated"
	UserDeletedEventType                    = "user.deleted"
	UserPreferenceUpdatedEventType          = "user.preference.updated"
)

// Tenant Domain Events

// TenantCreatedEvent is published when a new tenant is created
type TenantCreatedEvent struct {
	*events.BaseDomainEvent
	TenantName   string                 `json:"tenant_name"`
	TenantType   string                 `json:"tenant_type"`
	OwnerUserID  string                 `json:"owner_user_id"`
	Settings     map[string]interface{} `json:"settings"`
}

// NewTenantCreatedEvent creates a new tenant created event
func NewTenantCreatedEvent(
	tenantID, tenantName, tenantType, ownerUserID string,
	settings map[string]interface{},
) *TenantCreatedEvent {
	payload := map[string]interface{}{
		"tenant_name":   tenantName,
		"tenant_type":   tenantType,
		"owner_user_id": ownerUserID,
		"settings":      settings,
	}
		return &TenantCreatedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(TenantCreatedEventType, tenantID, tenantID, payload),
		TenantName:      tenantName,
		TenantType:      tenantType,
		OwnerUserID:     ownerUserID,
		Settings:        settings,
	}
}

// TenantConfigurationUpdatedEvent is published when tenant configuration changes
type TenantConfigurationUpdatedEvent struct {
	*events.BaseDomainEvent
	ConfigKey    string      `json:"config_key"`
	OldValue     interface{} `json:"old_value"`
	NewValue     interface{} `json:"new_value"`
	UpdatedBy    string      `json:"updated_by"`
}

// NewTenantConfigurationUpdatedEvent creates a new tenant configuration updated event
func NewTenantConfigurationUpdatedEvent(
	tenantID, configKey string,
	oldValue, newValue interface{},
	updatedBy string,
) *TenantConfigurationUpdatedEvent {
	payload := map[string]interface{}{
		"config_key": configKey,
		"old_value":  oldValue,
		"new_value":  newValue,
		"updated_by": updatedBy,
	}
		return &TenantConfigurationUpdatedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(TenantConfigurationUpdatedEventType, tenantID, tenantID, payload),
		ConfigKey:       configKey,
		OldValue:        oldValue,
		NewValue:        newValue,
		UpdatedBy:       updatedBy,
	}
}

// UserPreferenceUpdatedEvent is published when user preferences change
type UserPreferenceUpdatedEvent struct {
	*events.BaseDomainEvent
	UserID         string                 `json:"user_id"`
	PreferenceType string                 `json:"preference_type"`
	OldValue       interface{}            `json:"old_value"`
	NewValue       interface{}            `json:"new_value"`
	UpdatedBy      string                 `json:"updated_by"`
	Preferences    map[string]interface{} `json:"preferences"`
}

// NewUserPreferenceUpdatedEvent creates a new user preference updated event
func NewUserPreferenceUpdatedEvent(
	userID, tenantID, preferenceType string,
	oldValue, newValue interface{},
	updatedBy string,
	preferences map[string]interface{},
) *UserPreferenceUpdatedEvent {
	payload := map[string]interface{}{
		"user_id":         userID,
		"preference_type": preferenceType,
		"old_value":       oldValue,
		"new_value":       newValue,
		"updated_by":      updatedBy,
		"preferences":     preferences,
	}
		return &UserPreferenceUpdatedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(UserPreferenceUpdatedEventType, userID, tenantID, payload),
		UserID:          userID,
		PreferenceType:  preferenceType,
		OldValue:        oldValue,
		NewValue:        newValue,
		UpdatedBy:       updatedBy,
		Preferences:     preferences,
	}
}

// UserCreatedEvent is published when a new user is created
type UserCreatedEvent struct {
	*events.BaseDomainEvent
	UserID      string                 `json:"user_id"`
	Email       string                 `json:"email"`
	Name        string                 `json:"name"`
	Role        string                 `json:"role"`
	CreatedBy   string                 `json:"created_by"`
	UserData    map[string]interface{} `json:"user_data"`
}

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(
	userID, tenantID, email, name, role, createdBy string,
	userData map[string]interface{},
) *UserCreatedEvent {
	payload := map[string]interface{}{
		"user_id":    userID,
		"email":      email,
		"name":       name,
		"role":       role,
		"created_by": createdBy,
		"user_data":  userData,
	}
		return &UserCreatedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(UserCreatedEventType, userID, tenantID, payload),
		UserID:          userID,
		Email:           email,
		Name:            name,
		Role:            role,
		CreatedBy:       createdBy,
		UserData:        userData,
	}
}
