package createprovider

import (
	"context"
	"getnoti.com/internal/providers/domain"
	"getnoti.com/internal/providers/repos"
	"getnoti.com/internal/shared/utils"
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
		ID:          utils.GenerateUUID(),
		Name:        req.Name,
		Channels:    make([]domain.PrioritizedChannel, 0, len(req.Channels)),
		Credentials: req.Credentials,
	}

	// Set up channels with priorities
	for _, channelType := range req.Channels {
		// Get the next available priority for this channel
		priority, err := uc.repo.GetNextAvailablePriority(ctx, string(channelType))
		if err != nil {
			return CreateProviderResponse{}, err
		}

		provider.Channels = append(provider.Channels, domain.PrioritizedChannel{
			Type:     channelType,
			Priority: priority,
			Enabled:  true, // Assuming new channels are enabled by default
		})
	}

	// Create the provider in the repository
	createdProvider, err := uc.repo.CreateProvider(ctx, provider)
	if err != nil {
		return CreateProviderResponse{}, err
	}

	// Prepare the response
	channelDTOs := make([]ProviderChannelDTO, 0, len(createdProvider.Channels))
	for _, channel := range createdProvider.Channels {
		channelDTOs = append(channelDTOs, ProviderChannelDTO{
			Type:     channel.Type,
			Priority: channel.Priority,
			Enabled:  channel.Enabled,
		})
	}

	return CreateProviderResponse{
		ID:          createdProvider.ID,
		Name:        createdProvider.Name,
		Channels:    channelDTOs,
		Credentials: createdProvider.Credentials,
	}, nil
}
