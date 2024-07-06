package createtenant

import (
    "context"
    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
)

// Interface Definition
type CreateTenantUseCase interface {
    Execute(ctx context.Context, input CreateTenantInput) (CreateTenantOutput, error)
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
func (uc *createTenantUseCase) Execute(ctx context.Context, input CreateTenantInput) (CreateTenantOutput, error) {
    tenant := domain.Tenant{
        ID:             input.ID,
        Name:           input.Name,
        DefaultChannel: string(input.DefaultChannel),
        Preferences:    make(map[string]domain.ChannelPreference),
    }

    err := uc.repo.CreateTenant(ctx, tenant)
    if err != nil {
        return CreateTenantOutput{}, err
    }

    return CreateTenantOutput{Tenant: tenant}, nil
}
