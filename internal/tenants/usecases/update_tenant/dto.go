package updatetenant

import (
	"getnoti.com/internal/tenants/domain"
)

type UpdateTenantRequest struct {
	ID string
	Name           string
	DBConfigs map[string]*domain.DBCredentials
}

type UpdateTenantResponse struct {
	Success bool
	Tenant  domain.Tenant
	Error   string
}
