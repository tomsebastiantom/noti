// package providers

// import (
//     "context"
//     "fmt"
//     "getnoti.com/internal/providers/repos"
//     "getnoti.com/pkg/cache"
//     "getnoti.com/pkg/vault"
// )

// type ProviderFactory struct {
//     providerCache *cache.GenericCache
//     providerRepo  repos.ProviderRepository
  
// }

// func NewProviderFactory(providerCache *cache.GenericCache, providerRepo repos.ProviderRepository) *ProviderFactory {
//     return &ProviderFactory{
//         providerCache: providerCache,
//         providerRepo:  providerRepo,
    
//     }
// }

// func (f *ProviderFactory) GetProvider( providerID string, tenantID string, channel string) (Provider, error) {
//     key := providerID + ":" + tenantID + ":" + channel

//     if cachedProvider, exists := f.providerCache.Get(key); exists {
//         if provider, ok := cachedProvider.(Provider); ok {
//             return provider, nil
//         }
//     }

//     providerDTO, err := f.providerRepo.GetProviderByID(context.Background(), providerID)
//     if err != nil {
//         return nil, fmt.Errorf("failed to get provider: %v", err)
//     }

// 	var provider Provider
//     switch providerDTO.Name {e
//     case "twilio":
//         credentials, err := vault.GetClientCredentials(tenantID, vault.GenericCredential, providerDTO.Name)
//         if err != nil {
//             return nil, fmt.Errorf("failed to get client credentials: %v", err)
//         }
        
//         accountSid, ok := credentials["account_sid"].(string)
//         if !ok {
//             return nil, fmt.Errorf("invalid account_sid in credentials")
//         }
//         authToken, ok := credentials["auth_token"].(string)
//         if !ok {
//             return nil, fmt.Errorf("invalid auth_token in credentials")
//         }
        
//         provider = NewTwilioProvider(accountSid, authToken)
//     // ... (other providers)
//     }

//     // Cache the new provider
//     f.providerCache.Set(key, provider, 1)
//     return provider, nil
// }
package providers

import (
    "context"
    "fmt"
    "getnoti.com/internal/providers/repos"
    "getnoti.com/pkg/cache"
    "getnoti.com/pkg/credentials"
)

type ProviderFactory struct {
    providerCache     *cache.GenericCache
    providerRepo      repos.ProviderRepository
    credentialManager *credentials.Manager
}

func NewProviderFactory(providerCache *cache.GenericCache, providerRepo repos.ProviderRepository, credentialManager *credentials.Manager) *ProviderFactory {
    return &ProviderFactory{
        providerCache:     providerCache,
        providerRepo:      providerRepo,
        credentialManager: credentialManager,
    }
}

func (f *ProviderFactory) GetProvider(providerID string, tenantID string, channel string) (Provider, error) {
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
        credMap, err := f.credentialManager.GetCredentials(tenantID, credentials.GenericCredential, providerDTO.Name)
        if err != nil {
            return nil, fmt.Errorf("failed to get twilio credentials: %v", err)
        }
        
        accountSid, ok := credMap["account_sid"].(string)
        if !ok {
            return nil, fmt.Errorf("invalid account_sid in credentials")
        }
        authToken, ok := credMap["auth_token"].(string)
        if !ok {
            return nil, fmt.Errorf("invalid auth_token in credentials")
        }
        
        provider = NewTwilioProvider(accountSid, authToken)
        
    // case "aws_ses":
    //     credMap, err := f.credentialManager.GetCredentials(tenantID, credentials.GenericCredential, "aws")
    //     if err != nil {
    //         return nil, fmt.Errorf("failed to get AWS credentials: %v", err)
    //     }
        
    //     accessKey, ok := credMap["access_key"].(string)
    //     if !ok {
    //         return nil, fmt.Errorf("invalid AWS access_key in credentials")
    //     }
    //     secretKey, ok := credMap["secret_key"].(string)
    //     if !ok {
    //         return nil, fmt.Errorf("invalid AWS secret_key in credentials")
    //     }
    //     region, ok := credMap["region"].(string)
    //     if !ok {
    //         return nil, fmt.Errorf("invalid AWS region in credentials")
    //     }
        
    //     provider = NewAWSSESProvider(accessKey, secretKey, region)
        
    // case "sendgrid":
    //     credMap, err := f.credentialManager.GetCredentials(tenantID, credentials.GenericCredential, providerDTO.Name)
    //     if err != nil {
    //         return nil, fmt.Errorf("failed to get sendgrid credentials: %v", err)
    //     }
        
    //     apiKey, ok := credMap["api_key"].(string)
    //     if !ok {
    //         return nil, fmt.Errorf("invalid api_key in credentials")
    //     }
        
    //     provider = NewSendGridProvider(apiKey)
        
    default:
        return nil, fmt.Errorf("unsupported provider: %s", providerDTO.Name)
    }

    // Cache the new provider
    f.providerCache.Set(key, provider, 1)
    return provider, nil
}