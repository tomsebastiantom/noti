package createtenant

import (
    "context"
    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
)

// Interface Definition
type CreateTenantUseCase interface {
    Execute(ctx context.Context, input CreateTenantRequest) (CreateTenantResponse, error)
}

// Struct Implementation
type createTenantUseCase struct {
    repo repository.TenantRepository
}

func NewCreateTenantUseCase(repo repository.TenantRepository) CreateTenantUseCase {
    return &createTenantUseCase{
        repo: repo,
    }
}

// Method Implementation
func (uc *createTenantUseCase) Execute(ctx context.Context, req CreateTenantRequest) (CreateTenantResponse, error) {
    // Create a new tenant using the domain package
    tenant := domain.NewTenant(req.ID, req.Name)

    // Add DB configurations if provided
    for key, config := range req.DBConfigs {
        dbCreds, err := domain.NewDBCredentials(
            config.Type,
            config.DSN,
            config.Host,
            config.Port,
            config.Username,
            config.Password,
            config.DBName,
        )
        if err != nil {
            return CreateTenantResponse{}, err
        }
        tenant.AddDBConfig(key, dbCreds)
    }

    // Validate the tenant
    if err := tenant.Validate(); err != nil {
        return CreateTenantResponse{}, err
    }

    // Create tenant in the repository
    err := uc.repo.CreateTenant(ctx, *tenant)
    if err != nil {
        return CreateTenantResponse{}, err
    }

    return CreateTenantResponse{Tenant: *tenant}, nil
}