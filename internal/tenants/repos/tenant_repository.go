package repository

import (
	"context"
	"getnoti.com/internal/tenants/domain"
)

type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant domain.Tenant) error
	GetTenantByID(ctx context.Context, tenantID string) (domain.Tenant, error)
	Update(ctx context.Context, tenant domain.Tenant) error
	

}


type TenantsRepository interface {
	GetAllTenants(ctx context.Context) ([]domain.Tenant, error)
}
