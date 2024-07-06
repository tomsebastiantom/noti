
package updatetenant

import (
    "context"
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

func (uc *updateTenantUseCase) Execute(ctx context.Context, input UpdateTenantRequest) UpdateTenantResponse {
    tenant, err := uc.repo.GetTenantByID(ctx, input.ID)
    if err != nil {
        return UpdateTenantResponse{Success: false, Error: err.Error()}
    }

    tenant.Name = input.Name
    tenant.DefaultChannel = input.DefaultChannel
    tenant.Preferences = input.Preferences

    if err := uc.repo.Update(ctx, tenant); err != nil {
        return UpdateTenantResponse{Success: false, Error: err.Error()}
    }

    return UpdateTenantResponse{Success: true, Tenant: tenant}
}
