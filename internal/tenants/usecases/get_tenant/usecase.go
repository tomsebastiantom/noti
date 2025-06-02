package gettenant

import (
	"context"
	"getnoti.com/internal/tenants/repos"
)

type GetTenantUseCase interface {
	Execute(ctx context.Context, req GetTenantRequest) (GetTenantResponse, error)
}

type getTenantUseCase struct {
	repo repository.TenantsRepository
}

func NewGetTenantUseCase(repo repository.TenantsRepository) GetTenantUseCase {
	return &getTenantUseCase{
		repo: repo,
	}
}

func (uc *getTenantUseCase) Execute(ctx context.Context, req GetTenantRequest) (GetTenantResponse, error) {
	tenant, err := uc.repo.GetTenantByID(ctx, req.TenantID)
	if err != nil {
		return GetTenantResponse{}, err
	}
	return GetTenantResponse{
		ID:   tenant.ID,
		Name: tenant.Name}, nil

}
