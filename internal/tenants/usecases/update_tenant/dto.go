package updatetenant

import (
	"getnoti.com/internal/tenants/domain"
)

type UpdateTenantRequest struct {
	ID             string
	Name           string
	Preferences    map[string]domain.ChannelPreference
}

type UpdateTenantResponse struct {
	Success bool
	Tenant  domain.Tenant
	Error   string
}
