package createtenants

import (
    "context"
    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
)

type CreateTenantsUseCase interface {
    Execute(ctx context.Context, input CreateTenantsRequest) (CreateTenantsResponse, error)
}

type createTenantsUseCase struct {
    repo repository.TenantRepository
}

func NewCreateTenantsUseCase(repo repository.TenantRepository) CreateTenantsUseCase {
    return &createTenantsUseCase{
        repo: repo,
    }
}

func (uc *createTenantsUseCase) Execute(ctx context.Context, req CreateTenantsRequest) (CreateTenantsResponse, error) {
    var successTenants []string
    var failedTenants []FailedTenant

    for _, tenantReq := range req.Tenants {
        if tenantReq.ID == "" || tenantReq.Name == "" {
            failedTenants = append(failedTenants, FailedTenant{ID: tenantReq.ID, Error: ErrMissingRequiredFields.Error()})
            continue
        }

        tenant := domain.NewTenant(tenantReq.ID, tenantReq.Name)

        // Add DB configurations if provided
        for key, config := range tenantReq.DBConfigs {
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
                failedTenants = append(failedTenants, FailedTenant{ID: tenantReq.ID, Error: err.Error()})
                continue
            }
            tenant.AddDBConfig(key, dbCreds)
        }

        // Validate the tenant
        if err := tenant.Validate(); err != nil {
            failedTenants = append(failedTenants, FailedTenant{ID: tenantReq.ID, Error: err.Error()})
            continue
        }

        err := uc.repo.CreateTenant(ctx, *tenant)
        if err != nil {
            failedTenants = append(failedTenants, FailedTenant{ID: tenantReq.ID, Error: err.Error()})
        } else {
            successTenants = append(successTenants, tenantReq.ID)
        }
    }

    if len(successTenants) == 0 {
        return CreateTenantsResponse{}, ErrAllTenantsCreationFailed
    }

    return CreateTenantsResponse{
        SuccessTenants: successTenants,
        FailedTenants:  failedTenants,
    }, nil
}


