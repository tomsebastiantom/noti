// package vault

// // VaultProvider interface for different vault implementations
// type VaultProvider interface {
//     Initialize(config *VaultConfig) error
//     CreateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error
//     GetClientCredentials(tenantID string, credType CredentialType, name string) (map[string]interface{}, error)
//     UpdateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error
//     DeleteCredential(tenantID string, credType CredentialType, name string) error
//     IsHealthy() bool
//     GetProviderType() string
// }

// // VaultConfig structure
// type VaultConfig struct {
//     Address  string
//     Token    string
//     Provider string // "hashicorp", "aws", "azure", "gcp"
// }

// // CredentialType enum
// type CredentialType string

// const (
//     DBCredential      CredentialType = "db"
//     GenericCredential CredentialType = "generic"
// )

// // Credential structure
// type Credential struct {
//     Type        CredentialType         `json:"type"`
//     Identifier  string                 `json:"identifier,omitempty"`
//     Secret      string                 `json:"secret,omitempty"`
//     ExtraParams map[string]interface{} `json:"extra_params,omitempty"`
// }

// // Global vault provider instance
// var (
//     currentProvider VaultProvider
//     isInitialized   bool
// )

// // Initialize initializes the vault provider based on config
// func Initialize(config *VaultConfig) error {
//     switch config.Provider {
//     case "hashicorp", "":
//         currentProvider = NewHashiCorpProvider()
//     case "aws":
//         currentProvider = NewAWSSecretsManagerProvider()
//     case "azure":
//         currentProvider = NewAzureKeyVaultProvider()
//     case "gcp":
//         currentProvider = NewGCPSecretManagerProvider()
//     default:
//         currentProvider = NewHashiCorpProvider() // Default fallback
//     }
    
//     err := currentProvider.Initialize(config)
//     if err != nil {
//         return err
//     }
    
//     isInitialized = true
//     return nil
// }

// // CreateCredential creates a credential using the current provider
// func CreateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error {
//     if !isInitialized || currentProvider == nil {
//         return ErrVaultNotInitialized
//     }
//     return currentProvider.CreateCredential(tenantID, credType, name, data)
// }

// // GetClientCredentials gets credentials using the current provider
// func GetClientCredentials(tenantID string, credType CredentialType, name string) (map[string]interface{}, error) {
//     if !isInitialized || currentProvider == nil {
//         return nil, ErrVaultNotInitialized
//     }
//     return currentProvider.GetClientCredentials(tenantID, credType, name)
// }

// // UpdateCredential updates a credential using the current provider
// func UpdateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error {
//     if !isInitialized || currentProvider == nil {
//         return ErrVaultNotInitialized
//     }
//     return currentProvider.UpdateCredential(tenantID, credType, name, data)
// }

// // DeleteCredential deletes a credential using the current provider
// func DeleteCredential(tenantID string, credType CredentialType, name string) error {
//     if !isInitialized || currentProvider == nil {
//         return ErrVaultNotInitialized
//     }
//     return currentProvider.DeleteCredential(tenantID, credType, name)
// }

// // IsHealthy checks if the current provider is healthy
// func IsHealthy() bool {
//     if !isInitialized || currentProvider == nil {
//         return false
//     }
//     return currentProvider.IsHealthy()
// }

// // GetProviderType returns the type of the current provider
// func GetProviderType() string {
//     if !isInitialized || currentProvider == nil {
//         return "none"
//     }
//     return currentProvider.GetProviderType()
// }

// // ParseCredentials parses raw credentials into structured format
// func ParseCredentials(rawCredentials map[string]interface{}, credType CredentialType) (*Credential, error) {
//     data, ok := rawCredentials["data"].(map[string]interface{})
//     if !ok {
//         return nil, ErrInvalidCredentialFormat
//     }

//     cred := &Credential{
//         Type:        credType,
//         ExtraParams: make(map[string]interface{}),
//     }

//     for key, value := range data {
//         switch key {
//         case "identifier", "username":
//             if str, ok := value.(string); ok {
//                 cred.Identifier = str
//             }
//         case "secret", "password":
//             if str, ok := value.(string); ok {
//                 cred.Secret = str
//             }
//         default:
//             cred.ExtraParams[key] = value
//         }
//     }

//     return cred, nil


package vault

// VaultProvider interface for different vault implementations
type VaultProvider interface {
    Initialize(config *VaultConfig) error
    CreateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error
    GetClientCredentials(tenantID string, credType CredentialType, name string) (map[string]interface{}, error)
    UpdateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error
    DeleteCredential(tenantID string, credType CredentialType, name string) error
    IsHealthy() bool
    GetProviderType() string
}

// VaultConfig structure
type VaultConfig struct {
    Address  string
    Token    string
    Provider string // "hashicorp", "aws", "azure", "gcp"
}

// CredentialType enum
type CredentialType string

const (
    DBCredential      CredentialType = "db"
    GenericCredential CredentialType = "generic"
)

// Credential structure
type Credential struct {
    Type        CredentialType         `json:"type"`
    Identifier  string                 `json:"identifier,omitempty"`
    Secret      string                 `json:"secret,omitempty"`
    ExtraParams map[string]interface{} `json:"extra_params,omitempty"`
}

// Global vault provider instance
var (
    currentProvider VaultProvider
    isInitialized   bool
)

// Initialize initializes the vault provider based on config
func Initialize(config *VaultConfig) error {
    switch config.Provider {
    case "hashicorp", "":
        currentProvider = NewHashiCorpProvider()
    // case "aws":
    //     currentProvider = NewAWSSecretsManagerProvider()
    // case "azure":
    //     currentProvider = NewAzureKeyVaultProvider()
    // case "gcp":
    //     currentProvider = NewGCPSecretManagerProvider()
    default:
        currentProvider = NewHashiCorpProvider() // Default fallback
    }
    
    err := currentProvider.Initialize(config)
    if err != nil {
        return err
    }
    
    isInitialized = true
    return nil
}

// CreateCredential creates a credential using the current provider
func CreateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error {
    if !isInitialized || currentProvider == nil {
        return ErrVaultNotInitialized
    }
    return currentProvider.CreateCredential(tenantID, credType, name, data)
}

// GetClientCredentials gets credentials using the current provider
func GetClientCredentials(tenantID string, credType CredentialType, name string) (map[string]interface{}, error) {
    if !isInitialized || currentProvider == nil {
        return nil, ErrVaultNotInitialized
    }
    return currentProvider.GetClientCredentials(tenantID, credType, name)
}

// UpdateCredential updates a credential using the current provider
func UpdateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error {
    if !isInitialized || currentProvider == nil {
        return ErrVaultNotInitialized
    }
    return currentProvider.UpdateCredential(tenantID, credType, name, data)
}

// DeleteCredential deletes a credential using the current provider
func DeleteCredential(tenantID string, credType CredentialType, name string) error {
    if !isInitialized || currentProvider == nil {
        return ErrVaultNotInitialized
    }
    return currentProvider.DeleteCredential(tenantID, credType, name)
}

// IsHealthy checks if the current provider is healthy
func IsHealthy() bool {
    if !isInitialized || currentProvider == nil {
        return false
    }
    return currentProvider.IsHealthy()
}

// GetProviderType returns the type of the current provider
func GetProviderType() string {
    if !isInitialized || currentProvider == nil {
        return "none"
    }
    return currentProvider.GetProviderType()
}

// ParseCredentials parses raw credentials into structured format
func ParseCredentials(rawCredentials map[string]interface{}, credType CredentialType) (*Credential, error) {
    data, ok := rawCredentials["data"].(map[string]interface{})
    if !ok {
        return nil, ErrInvalidCredentialFormat
    }

    cred := &Credential{
        Type:        credType,
        ExtraParams: make(map[string]interface{}),
    }

    for key, value := range data {
        switch key {
        case "identifier", "username":
            if str, ok := value.(string); ok {
                cred.Identifier = str
            }
        case "secret", "password":
            if str, ok := value.(string); ok {
                cred.Secret = str
            }
        default:
            cred.ExtraParams[key] = value
        }
    }

    return cred, nil
}