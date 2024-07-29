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
	// Create a new Provider
	provider := &domain.Provider{
		Name:     req.Name,
		Enabled:  true, // Assuming new providers are enabled by default
		Channels: make(map[string]domain.ProviderChannel),
	}

	// Set up channels with priorities
	for _, channelName := range req.Channels {
		// Get the next available priority for this channel
		priority, err := uc.repo.GetNextAvailablePriority(ctx, channelName)
		if err != nil {
			return CreateProviderResponse{}, err
		}

		// If no priority is available in the database, set it to the highest priority (1)
		if priority == 0 {
			priority = 1
		}

		provider.Channels[channelName] = domain.ProviderChannel{
			Channel:  channelName,
			Priority: priority,
		}
	}

	// Create the provider in the repository
	createdProvider, err := uc.repo.CreateProvider(ctx, provider)
	if err != nil {
		return CreateProviderResponse{}, err
	}

	// Prepare the response
	channelDTOs := make([]ProviderChannelDTO, 0, len(createdProvider.Channels))
	for channelName, channel := range createdProvider.Channels {
		channelDTOs = append(channelDTOs, ProviderChannelDTO{
			Channel:  channelName,
			Priority: channel.Priority,
		})
	}

	return CreateProviderResponse{
		ID:       createdProvider.ID,
		Name:     createdProvider.Name,
		Channels: channelDTOs,
		Enabled:  createdProvider.Enabled,
	}, nil
}
