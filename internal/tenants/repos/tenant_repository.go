package repository

import (
	"context"
	"getnoti.com/internal/tenants/domain"
)

type TenantsRepository interface {
	CreateTenant(ctx context.Context, tenant domain.Tenant) error
	GetTenantByID(ctx context.Context, tenantID string) (domain.Tenant, error)
	Update(ctx context.Context, tenant domain.Tenant) error
    GetAllTenants(ctx context.Context) ([]domain.Tenant, error)

}

