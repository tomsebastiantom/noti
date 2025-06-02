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
    repo repository.TenantsRepository
}

func NewCreateTenantUseCase(repo repository.TenantsRepository) CreateTenantUseCase {
    return &createTenantUseCase{
        repo: repo,
    }
}

// Method Implementation
func (uc *createTenantUseCase) Execute(ctx context.Context, req CreateTenantRequest) (CreateTenantResponse, error) {
    // Create a new tenant using the domain package
    tenant := domain.NewTenant(req.ID, req.Name)

    // Add DB configurations if provided
    if req.DBConfig != nil {
        dbCreds, err := domain.NewDBCredentials(
            req.DBConfig.Type,
            req.DBConfig.DSN,
            req.DBConfig.Host,
            req.DBConfig.Port,
            req.DBConfig.Username,
            req.DBConfig.Password,
            req.DBConfig.DBName,
        )
        if err != nil {
            return CreateTenantResponse{}, err
        }
        tenant.SetDBConfig(dbCreds)
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

    return CreateTenantResponse{
        ID:   tenant.ID,
        Name: tenant.Name,
    }, nil
}