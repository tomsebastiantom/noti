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
        if tenantReq.DBConfig != nil {
            dbCreds, err := domain.NewDBCredentials(
                tenantReq.DBConfig.Type,
                tenantReq.DBConfig.DSN,
                tenantReq.DBConfig.Host,
                tenantReq.DBConfig.Port,
                tenantReq.DBConfig.Username,
                tenantReq.DBConfig.Password,
                tenantReq.DBConfig.DBName,
            )
            if err != nil {
                failedTenants = append(failedTenants, FailedTenant{ID: tenantReq.ID, Error: err.Error()})
                continue
            }
            tenant.SetDBConfig(dbCreds)
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


