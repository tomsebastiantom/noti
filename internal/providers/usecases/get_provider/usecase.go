package getprovider

import (
    "context"
    "getnoti.com/internal/providers/repos"
)

type GetProviderUseCase interface {
    Execute(ctx context.Context, req GetProviderRequest) (GetProviderResponse, error)
}

type getProviderUseCase struct {
    repo repos.ProviderRepository
}

func NewGetProviderUseCase(repo repos.ProviderRepository) GetProviderUseCase {
    return &getProviderUseCase{
        repo: repo,
    }
}

func (uc *getProviderUseCase) Execute(ctx context.Context, req GetProviderRequest) (GetProviderResponse, error) {
    provider, err := uc.repo.GetProviderByID(ctx, req.ID)
    if err != nil {
        return GetProviderResponse{}, err
    }

    return GetProviderResponse{
        ID:       provider.ID,
        Name:     provider.Name,
        Channels: provider.Channels,
        TenantID: provider.TenantID,
    }, nil
}