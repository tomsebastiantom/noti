package getuserpreferences

import (
	"getnoti.com/internal/tenants/domain"
)

type GetUserPreferencesRequest struct {
	TenantID string
	UserID   string
}

func (r *GetUserPreferencesRequest) SetTenantID(id string) {
	r.TenantID = id
}

type GetUserPreferencesResponse struct {
	ID             string                                  `json:"id"`
	UserID         string                                  `json:"userId"`
	Enabled        bool                                    `json:"enabled"`
	ChannelPrefs   map[domain.ChannelType]bool             `json:"channelPrefs"`
	CategoryPrefs  map[string]domain.CategoryPreference    `json:"categoryPrefs"`
	DigestSettings domain.DigestSettings                   `json:"digestSettings"`
}

// FromDomain converts domain UserPreference to response DTO
func FromDomain(pref domain.UserPreference) GetUserPreferencesResponse {
	return GetUserPreferencesResponse{
		ID:             pref.ID,
		UserID:         pref.UserID,
		Enabled:        pref.Enabled,
		ChannelPrefs:   pref.ChannelPrefs,
		CategoryPrefs:  pref.CategoryPrefs,
		DigestSettings: pref.DigestSettings,
	}
}
