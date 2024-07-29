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

    channelDTOs := make([]ProviderChannelDTO, 0, len(provider.Channels))
    for _, channel := range provider.Channels {
        channelDTOs = append(channelDTOs, ProviderChannelDTO{
            Channel:  channel.Channel,
            Priority: channel.Priority,
        })
    }

    return GetProviderResponse{
        ID:       provider.ID,
        Name:     provider.Name,
        Channels: channelDTOs,
        Enabled:  provider.Enabled,
    }, nil
}
