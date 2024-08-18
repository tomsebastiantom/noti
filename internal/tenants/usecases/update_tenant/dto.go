package updatetenant

import (
	"getnoti.com/internal/tenants/domain"
)

type UpdateTenantRequest struct {
	ID string
	Name           string
	DBConfig *domain.DBCredentials
}

type UpdateTenantResponse struct {
	Success bool
		ID string
	Name           string
}
