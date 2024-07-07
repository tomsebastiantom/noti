package providers

import (
    // "context"
    "sync"
    // "getnoti.com/internal/providers/dtos"
    "github.com/twilio/twilio-go"
)

type ProviderFactory struct {
    mu       sync.Mutex
    clients  map[string]Provider
}

func NewProviderFactory() *ProviderFactory {
    return &ProviderFactory{
        clients: make(map[string]Provider),
    }
}

func (f *ProviderFactory) GetProvider(providerID string, tenantID string, channel string) Provider {
    f.mu.Lock()
    defer f.mu.Unlock()

    key := providerID + ":" + tenantID + ":" + channel
    if provider, exists := f.clients[key]; exists {
        return provider
    }

    var provider Provider
    switch providerID {
    case "twilio":
        client := twilio.NewRestClientWithParams(twilio.ClientParams{})
        provider = NewTwilioProviderWithClient(client)
    // case "amazon_ses":
    //     provider = NewAmazonSESProvider()
    // case "mailgun":
    //     provider = NewMailgunProvider()
    // Add more providers as needed
    default:
        return nil
    }

    f.clients[key] = provider
    return provider
}
