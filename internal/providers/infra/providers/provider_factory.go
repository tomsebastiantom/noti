package providers

import (
    "context"
    "fmt"
    "getnoti.com/internal/providers/repos"
    "getnoti.com/pkg/cache"
    "getnoti.com/pkg/vault"
)

type ProviderFactory struct {
    providerCache *cache.GenericCache
    providerRepo  repos.ProviderRepository
  
}

func NewProviderFactory(providerCache *cache.GenericCache, providerRepo repos.ProviderRepository) *ProviderFactory {
    return &ProviderFactory{
        providerCache: providerCache,
        providerRepo:  providerRepo,
    
    }
}

func (f *ProviderFactory) GetProvider( providerID string, tenantID string, channel string) (Provider, error) {
    key := providerID + ":" + tenantID + ":" + channel

    if cachedProvider, exists := f.providerCache.Get(key); exists {
        if provider, ok := cachedProvider.(Provider); ok {
            return provider, nil
        }
    }

    providerDTO, err := f.providerRepo.GetProviderByID(context.Background(), providerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get provider: %v", err)
    }

	var provider Provider
    switch providerDTO.Name {
    case "twilio":
        credentials, err := vault.GetClientCredentials(tenantID, vault.GenericCredential, providerDTO.Name)
        if err != nil {
            return nil, fmt.Errorf("failed to get client credentials: %v", err)
        }
        
        accountSid, ok := credentials["account_sid"].(string)
        if !ok {
            return nil, fmt.Errorf("invalid account_sid in credentials")
        }
        authToken, ok := credentials["auth_token"].(string)
        if !ok {
            return nil, fmt.Errorf("invalid auth_token in credentials")
        }
        
        provider = NewTwilioProvider(accountSid, authToken)
    // ... (other providers)
    }

    // Cache the new provider
    f.providerCache.Set(key, provider, 1)
    return provider, nil
}
