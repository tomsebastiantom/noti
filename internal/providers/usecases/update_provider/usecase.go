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
    provider.Enabled = req.Enabled
    provider.Channels = make(map[string]domain.ProviderChannel)

    for _, channelDTO := range req.Channels {
        provider.Channels[channelDTO.Channel] = domain.ProviderChannel{
            Channel:  channelDTO.Channel,
            Priority: channelDTO.Priority,
        }
    }

    updatedProvider, err := uc.repo.UpdateProvider(ctx, provider)
    if err != nil {
        return UpdateProviderResponse{}, err
    }

    responseChannels := make([]ProviderChannelDTO, 0, len(updatedProvider.Channels))
    for _, channel := range updatedProvider.Channels {
        responseChannels = append(responseChannels, ProviderChannelDTO{
            Channel:  channel.Channel,
            Priority: channel.Priority,
        })
    }

    return UpdateProviderResponse{
        ID:       updatedProvider.ID,
        Name:     updatedProvider.Name,
        Channels: responseChannels,
        Enabled:  updatedProvider.Enabled,
    }, nil
}