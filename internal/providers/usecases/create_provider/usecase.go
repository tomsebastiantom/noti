package createprovider

import (
    "context"
    "getnoti.com/internal/providers/domain"
    "getnoti.com/internal/providers/repos"
)

type CreateProviderUseCase interface {
    Execute(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error)
}

type createProviderUseCase struct {
    repo repos.ProviderRepository
}

func NewCreateProviderUseCase(repo repos.ProviderRepository) CreateProviderUseCase {
    return &createProviderUseCase{
        repo: repo,
    }
}

func (uc *createProviderUseCase) Execute(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error) {
    // Implement business logic
    provider := &domain.Provider{
        Name:     req.Name,
        Channels: req.Channels,
        TenantID: req.TenantID,
    }

    createdProvider, err := uc.repo.CreateProvider(ctx, provider)
    if err != nil {
        return CreateProviderResponse{}, err
    }

    return CreateProviderResponse{
        ID:       createdProvider.ID,
        Name:     createdProvider.Name,
        Channels: createdProvider.Channels,
        TenantID: createdProvider.TenantID,
    }, nil
}