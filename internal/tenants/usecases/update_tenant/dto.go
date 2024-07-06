package updatetenant

import (
	"getnoti.com/internal/tenants/domain"
)

type UpdateTenantInput struct {
	ID             string
	Name           string
	DefaultChannel string
	Preferences    map[string]domain.ChannelPreference
}

type UpdateTenantOutput struct {
	Success bool
	Tenant  domain.Tenant
	Error   string
}
