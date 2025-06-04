package updateuserpreferences

import (
	"getnoti.com/internal/tenants/domain"
)

type UpdateUserPreferencesRequest struct {
	UserID         string                                  `json:"userId"`
	TenantID       string                                  `json:"tenantId"`
	Enabled        bool                                    `json:"enabled"`
	ChannelPrefs   map[domain.ChannelType]bool             `json:"channelPrefs,omitempty"`
	CategoryPrefs  map[string]domain.CategoryPreference    `json:"categoryPrefs,omitempty"`
	DigestSettings *domain.DigestSettings                  `json:"digestSettings,omitempty"`
}

func (r *UpdateUserPreferencesRequest) SetTenantID(id string) {
	r.TenantID = id
}

type UpdateUserPreferencesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
