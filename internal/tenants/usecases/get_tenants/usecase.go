package gettenants

import (
    "context"
    "getnoti.com/internal/tenants/repos"
)

type GetTenantsUseCase interface {
    Execute(ctx context.Context, req GetTenantsRequest) (GetTenantsResponse, error)
}

type getTenantsUseCase struct {
    repo repository.TenantsRepository
}

func NewGetTenantsUseCase(repo repository.TenantsRepository) GetTenantsUseCase {
    return &getTenantsUseCase{
        repo: repo,
    }
}

func (uc *getTenantsUseCase) Execute(ctx context.Context, req GetTenantsRequest) (GetTenantsResponse, error) {
   

    tenants, err := uc.repo.GetAllTenants(ctx)
    if err != nil {
        return GetTenantsResponse{}, err
    }
    return GetTenantsResponse{Tenants: tenants}, nil
}
