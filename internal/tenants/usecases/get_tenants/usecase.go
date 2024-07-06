package gettenants

import (
    "context"
    "getnoti.com/internal/tenants/domain"
    "getnoti.com/internal/tenants/repos"
)

type GetTenantsUseCase interface {
    Execute(ctx context.Context, req GetTenantsRequest) (GetTenantsResponse, error)
}

type getTenantsUseCase struct {
    repo repository.TenantRepository
}

func NewGetTenantsUseCase(repo repository.TenantRepository) GetTenantsUseCase {
    return &getTenantsUseCase{
        repo: repo,
    }
}

func (uc *getTenantsUseCase) Execute(ctx context.Context, req GetTenantsRequest) (GetTenantsResponse, error) {
    if req.TenantID != "" {
        tenant, err := uc.repo.GetTenantByID(ctx, req.TenantID)
        if err != nil {
            return GetTenantsResponse{}, err
        }
        return GetTenantsResponse{Tenants: []domain.Tenant{tenant}}, nil
    }

    tenants, err := uc.repo.GetAllTenants(ctx)
    if err != nil {
        return GetTenantsResponse{}, err
    }
    return GetTenantsResponse{Tenants: tenants}, nil
}
