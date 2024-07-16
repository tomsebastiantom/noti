package getproviderbytenant

import (
    "context"
    "getnoti.com/internal/providers/repos"
)

type GetProviderByTenantUseCase interface {
    Execute(ctx context.Context, req GetProviderByTenantRequest) (GetProviderByTenantResponse, error)
}

type getProviderByTenantUseCase struct {
    repo repos.ProviderRepository
}

func NewGetProviderByTenantUseCase(repo repos.ProviderRepository) GetProviderByTenantUseCase {
    return &getProviderByTenantUseCase{
        repo: repo,
    }
}

func (uc *getProviderByTenantUseCase) Execute(ctx context.Context, req GetProviderByTenantRequest) (GetProviderByTenantResponse, error) {
    providers, err := uc.repo.GetProvidersByTenantID(ctx, req.TenantID)
    if err != nil {
        return GetProviderByTenantResponse{}, err
    }

    var providerResponses []ProviderResponse
    for _, provider := range providers {
        providerResponses = append(providerResponses, ProviderResponse{
            ID:       provider.ID,
            Name:     provider.Name,
            Channels: provider.Channels,
            TenantID: provider.TenantID,
        })
    }

    return GetProviderByTenantResponse{
        Providers: providerResponses,
    }, nil
}