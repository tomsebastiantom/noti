package providers

import (

    // "getnoti.com/internal/providers/dtos"
	"getnoti.com/pkg/cache"
    "github.com/twilio/twilio-go"
)

type ProviderFactory struct {
    providerCache *cache.GenericCache  
}

func NewProviderFactory(providerCache *cache.GenericCache) *ProviderFactory {
    return &ProviderFactory{
        providerCache: providerCache,
    }
}

func (f *ProviderFactory) GetProvider(providerID string, tenantID string, channel string) Provider {
    key := providerID + ":" + tenantID + ":" + channel
    
    if cachedProvider, exists := f.providerCache.Get(key); exists {
        if provider, ok := cachedProvider.(Provider); ok {
            return provider
        }
    }

    var provider Provider
    switch providerID {
    case "twilio":
        client := twilio.NewRestClientWithParams(twilio.ClientParams{})
        provider = NewTwilioProviderWithClient(client)
    // Add more providers as needed
    default:
        return nil
    }

    // Cache the new provider
    f.providerCache.Set(key, provider,1) 
    return provider
}
