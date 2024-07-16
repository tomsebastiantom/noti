
package repos

import (
    "context"
    "getnoti.com/internal/providers/domain"
)

type ProviderRepository interface {
	CreateProvider(ctx context.Context, provider *domain.Provider) (*domain.Provider, error)
    GetProviderByID(ctx context.Context, id string) (*domain.Provider, error)
	GetProvidersByTenantID(ctx context.Context, tenantID string) ([]*domain.Provider, error)
    UpdateProvider(ctx context.Context, provider *domain.Provider) (*domain.Provider, error)
    
}
