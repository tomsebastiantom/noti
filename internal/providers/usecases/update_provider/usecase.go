package updateprovider

import (
    "context"
    "getnoti.com/internal/providers/repos"
)

type UpdateProviderUseCase interface {
    Execute(ctx context.Context, req UpdateProviderRequest) (UpdateProviderResponse, error)
}

type updateProviderUseCase struct {
    repo repos.ProviderRepository
}

func NewUpdateProviderUseCase(repo repos.ProviderRepository) UpdateProviderUseCase {
    return &updateProviderUseCase{
        repo: repo,
    }
}

func (uc *updateProviderUseCase) Execute(ctx context.Context, req UpdateProviderRequest) (UpdateProviderResponse, error) {
    provider, err := uc.repo.GetProviderByID(ctx, req.ID)
    if err != nil {
        return UpdateProviderResponse{}, err
    }

    provider.Name = req.Name
    provider.Channels = req.Channels

    updatedProvider, err := uc.repo.UpdateProvider(ctx, provider)
    if err != nil {
        return UpdateProviderResponse{}, err
    }

    return UpdateProviderResponse{
        ID:       updatedProvider.ID,
        Name:     updatedProvider.Name,
        Channels: updatedProvider.Channels,
        TenantID: updatedProvider.TenantID,
    }, nil
}