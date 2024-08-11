package vault

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
)

var (
	client *api.Client
	config *VaultConfig
	mutex  sync.RWMutex
	once   sync.Once
)

type VaultConfig struct {
	Address string
	Token   string
}

type CredentialType string

const (
	DBCredential      CredentialType = "db"
	GenericCredential CredentialType = "generic"
)

type Credential struct {
	Type        CredentialType         `json:"type"`
	Identifier  string                 `json:"identifier,omitempty"`
	Secret      string                 `json:"secret,omitempty"`
	ExtraParams map[string]interface{} `json:"extra_params,omitempty"`
}

func Initialize(cfg *VaultConfig) error {
    var err error
    once.Do(func() {
        fmt.Println("Initializing Vault...")

        if cfg == nil {
            err = fmt.Errorf("vault configuration is nil")
            fmt.Println("Error: Vault configuration is nil")
            return
        }

        fmt.Printf("Vault Address: %s, Token: %s\n", cfg.Address, cfg.Token)

        // Initialize the config variable
        config = &VaultConfig{
            Address: cfg.Address,
            Token:   cfg.Token,
        }

        fmt.Println("Creating Vault client...")
        vaultConfig := api.DefaultConfig()
        vaultConfig.Address = config.Address
        client, err = api.NewClient(vaultConfig)
        if err != nil {
            err = fmt.Errorf("failed to create Vault client: %v", err)
            fmt.Printf("Error creating Vault client: %v\n", err)
            return
        }

        fmt.Println("Setting Vault token...")
        client.SetToken(config.Token)

        fmt.Println("Vault initialization completed successfully")
    })
    return err
}

func ensureKVEngineMounted(client *api.Client, mountPath string) error {
	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return fmt.Errorf("failed to list mounts: %v", err)
	}

	if _, ok := mounts[mountPath]; !ok {
		// Mount the KV secrets engine
		err = client.Sys().Mount(mountPath, &api.MountInput{
			Type: "kv",
			Options: map[string]string{
				"version": "2",
			},
		})
		if err != nil {
			return fmt.Errorf("failed to mount KV engine: %v", err)
		}
		fmt.Println("KV secrets engine mounted successfully")
	} else {
		fmt.Println("KV secrets engine already mounted")
	}

	return nil
}


func refreshToken(tenantID string) error {
    mutex.Lock()
    defer mutex.Unlock()

    secret, err := client.Auth().Token().LookupSelf()
    if err != nil {
        return fmt.Errorf("failed to lookup token: %v", err)
    }

    // Check if the token has an expiration time
    expireTime, ok := secret.Data["expire_time"].(string)
    if !ok {
        // If there's no expiration time, assume it's a root token or a token without expiry
       // log.Println("Token does not have an expiration time, assuming it is a root token or a token without expiry.")
        return nil
    }

    expireTimeParsed, err := time.Parse(time.RFC3339, expireTime)
    if err != nil {
        return fmt.Errorf("failed to parse token expiration time: %v", err)
    }

    // Refresh the token if it will expire in less than 10 minutes
    if time.Until(expireTimeParsed) < 10*time.Minute {
        newToken, err := createTenantToken(tenantID)
        if err != nil {
            return fmt.Errorf("failed to refresh token: %v", err)
        }

        client.SetToken(newToken)
        
    } 

    return nil
}

func createTenantToken(tenantID string) (string, error) {
	policyName := fmt.Sprintf("%s-policy", tenantID)
	tokenRequest := &api.TokenCreateRequest{
		Policies: []string{policyName},
		TTL:      "1h",
	}

	secret, err := client.Auth().Token().Create(tokenRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant token: %v", err)
	}

	return secret.Auth.ClientToken, nil
}

func CreateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error {
	err := ensureTenantPolicy(tenantID)
	if err != nil {
		return err
	}

	err = refreshToken(tenantID)
	if err != nil {
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()

	secretPath := buildSecretPath(tenantID, credType, name)

	// Ensure the KV secrets engine is mounted
	err = ensureKVEngineMounted(client, "secret/")
	if err != nil {
		return err
	}

	_, err = client.Logical().Write(secretPath, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("failed to create credential: %v", err)
	}

	return nil
}



func GetClientCredentials(tenantID string, credType CredentialType, name string) (map[string]interface{}, error) {
	err := refreshToken(tenantID)
	if err != nil {
		return nil, err
	}

	mutex.RLock()
	defer mutex.RUnlock()

	secretPath := buildSecretPath(tenantID, credType, name)
	secret, err := client.Logical().Read(secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %v", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path: %s", secretPath)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid credential data format")
	}

	return data, nil
}

func buildSecretPath(tenantID string, credType CredentialType, name string) string {
	basePath := fmt.Sprintf("secret/data/tenants/%s/credentials/%s", tenantID, credType)
	if credType == DBCredential {
		return basePath
	}
	return fmt.Sprintf("%s/%s", basePath, name)
}

func ParseCredentials(rawCredentials map[string]interface{}, credType CredentialType) (*Credential, error) {
	data, ok := rawCredentials["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid credential data format")
	}

	cred := &Credential{
		Type:        credType,
		ExtraParams: make(map[string]interface{}),
	}

	for key, value := range data {
		switch key {
		case "identifier", "username":
			cred.Identifier = value.(string)
		case "secret", "password":
			cred.Secret = value.(string)
		default:
			cred.ExtraParams[key] = value
		}
	}

	return cred, nil
}

func UpdateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error {
	err := ensureTenantPolicy(tenantID)
	if err != nil {
		return err
	}

	err = refreshToken(tenantID)
	if err != nil {
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()

	secretPath := buildSecretPath(tenantID, credType, name)
	_, err = client.Logical().Write(secretPath, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("failed to update credential: %v", err)
	}

	return nil
}

func ensureTenantPolicy(tenantID string) error {
	mutex.Lock()
	defer mutex.Unlock()

	policyName := fmt.Sprintf("%s-policy", tenantID)
	policy := fmt.Sprintf(`
path "secret/data/tenants/%s/*" {
  capabilities = ["create", "read", "update", "delete"]
}
`, tenantID)

	err := client.Sys().PutPolicy(policyName, policy)
	if err != nil {
		return fmt.Errorf("failed to create policy: %v", err)
	}

	return nil
}
