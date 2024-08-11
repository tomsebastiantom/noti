package updatetenant

import (
	"context"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
)

type UpdateTenantUseCase interface {
    Execute(ctx context.Context, input UpdateTenantRequest) (UpdateTenantResponse)
}



type updateTenantUseCase struct {
    repo repository.TenantRepository
}

func NewUpdateTenantUseCase(repo repository.TenantRepository) UpdateTenantUseCase {
    return &updateTenantUseCase{
        repo: repo,
    }
}

func (uc *updateTenantUseCase) Execute(ctx context.Context, req UpdateTenantRequest) UpdateTenantResponse {
    tenant, err := uc.repo.GetTenantByID(ctx, req.ID)
    if err != nil {
        return UpdateTenantResponse{Success: false, Error: err.Error()}
    }

    // Update tenant name
    tenant.Name = req.Name

    // Update DB configurations
    if req.DBConfigs != nil {
        dbCreds, err := domain.NewDBCredentials(
            req.DBConfigs.Type,
            req.DBConfigs.DSN,
            req.DBConfigs.Host,
            req.DBConfigs.Port,
            req.DBConfigs.Username,
            req.DBConfigs.Password,
            req.DBConfigs.DBName,
        )
        if err != nil {
            return UpdateTenantResponse{Success: false, Error: err.Error()}
        }
        tenant.SetDBConfig(dbCreds)
    }

    // Validate the updated tenant
    if err := tenant.Validate(); err != nil {
        return UpdateTenantResponse{Success: false, Error: err.Error()}
    }

    if err := uc.repo.Update(ctx, tenant); err != nil {
        return UpdateTenantResponse{Success: false, Error: err.Error()}
    }

    return UpdateTenantResponse{Success: true, Tenant: tenant}
}

