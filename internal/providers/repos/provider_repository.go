
package repos

import (
    "context"
    "getnoti.com/internal/providers/domain"
)

type ProviderRepository interface {
	CreateProvider(ctx context.Context, provider *domain.Provider) (*domain.Provider, error)
    GetProviderByID(ctx context.Context, id string) (*domain.Provider, error)
    GetProviderByChannel(ctx context.Context, channelName string)(*domain.Provider, error)
    GetProvidersByChannel(ctx context.Context, channelName string) ([]*domain.Provider, error)
	GetProviders(ctx context.Context) ([]*domain.Provider, error)
    UpdateProvider(ctx context.Context, provider *domain.Provider) (*domain.Provider, error)
    GetNextAvailablePriority(ctx context.Context, channelName string) (int, error)
    
}
