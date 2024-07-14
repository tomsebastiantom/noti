package repository

import (
    "context"
    "getnoti.com/internal/tenants/domain"
    // "getnoti.com/pkg/vault"
)

type TenantRepository interface {
    CreateTenant(ctx context.Context, tenant domain.Tenant) error
    GetTenantByID(ctx context.Context, tenantID string) (domain.Tenant, error)
    Update(ctx context.Context, tenant domain.Tenant) error
    GetAllTenants(ctx context.Context) ([]domain.Tenant, error)
    GetPreferenceByChannel(ctx context.Context, tenantID string, channel string) (map[string]string, error)
    
    // New methods for Vault integration
   // StoreDBCredentialsInVault(ctx context.Context, tenantID string, dbConfig *domain.DBConfig) error
    // GetDBCredentialsFromVault(ctx context.Context, tenantID string) (*domain.DBConfig, error)
}

type TenantPreferenceRepository interface {
 
    GetPreferenceByChannel(ctx context.Context, tenantID string, channel string) (map[string]string, error)
    

}