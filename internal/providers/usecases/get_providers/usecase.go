package getproviders

import (
    "context"
    "getnoti.com/internal/providers/repos"
)

type GetProvidersUseCase interface {
    Execute(ctx context.Context, req GetProvidersRequest) (GetProvidersResponse, error)
}

type getProvidersUseCase struct {
    repo repos.ProviderRepository
}

func NewGetProvidersUseCase(repo repos.ProviderRepository) GetProvidersUseCase {
    return &getProvidersUseCase{
        repo: repo,
    }
}


func (uc *getProvidersUseCase) Execute(ctx context.Context, req GetProvidersRequest) (GetProvidersResponse, error) {
    providers, err := uc.repo.GetProviders(ctx)
    if err != nil {
        return GetProvidersResponse{}, err
    }

    providerResponses := make([]ProviderResponse, len(providers))
    for i, provider := range providers {
        channelDTOs := make([]ProviderChannelDTO, len(provider.Channels))
        for j, channel := range provider.Channels {
            channelDTOs[j] = ProviderChannelDTO{
                Type:     channel.Type,
                Priority: channel.Priority,
                Enabled:  channel.Enabled,
            }
        }

        providerResponses[i] = ProviderResponse{
            ID:          provider.ID,
            Name:        provider.Name,
            Channels:    channelDTOs,
            Credentials: provider.Credentials,
        }
    }

    return GetProvidersResponse{
        Providers: providerResponses,
    }, nil
}
