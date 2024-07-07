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
func (uc *createTenantUseCase) Execute(ctx context.Context, input CreateTenantRequest) (CreateTenantResponse, error) {
    // Map CreateTenantRequest to domain.Tenant
    preferences := make(map[string]domain.ChannelPreference)
    for key, pref := range input.Preferences {
        preferences[key] = domain.ChannelPreference{
            ChannelName: pref.ChannelName,
            Enabled:     pref.Enabled,
            ProviderID:  pref.ProviderID,
        }
    }

    tenant := domain.Tenant{
        ID:          input.ID,
        Name:        input.Name,
        Preferences: preferences,
    }

    // Create tenant in the repository
    err := uc.repo.CreateTenant(ctx, tenant)
    if err != nil {
        return CreateTenantResponse{}, err
    }

    return CreateTenantResponse{Tenant: tenant}, nil
}