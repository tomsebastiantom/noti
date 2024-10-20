package updateprovider

import (
    "context"
    "getnoti.com/internal/providers/repos"
    "getnoti.com/internal/providers/domain"
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
    provider.Credentials = req.Credentials
    provider.Channels = make([]domain.PrioritizedChannel, len(req.Channels))

    for i, channelDTO := range req.Channels {
        provider.Channels[i] = domain.PrioritizedChannel{
            Type:     channelDTO.Type,
            Priority: channelDTO.Priority,
            Enabled:  channelDTO.Enabled,
        }
    }

    updatedProvider, err := uc.repo.UpdateProvider(ctx, provider)
    if err != nil {
        return UpdateProviderResponse{}, err
    }

    responseChannels := make([]ProviderChannelDTO, len(updatedProvider.Channels))
    for i, channel := range updatedProvider.Channels {
        responseChannels[i] = ProviderChannelDTO{
            Type:     channel.Type,
            Priority: channel.Priority,
            Enabled:  channel.Enabled,
        }
    }

    return UpdateProviderResponse{
        ID:          updatedProvider.ID,
        Name:        updatedProvider.Name,
        Channels:    responseChannels,
        Credentials: updatedProvider.Credentials,
    }, nil
}