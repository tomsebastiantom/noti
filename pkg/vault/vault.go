package vault

import (
	"fmt"
	"sync"
	"time"
	"github.com/hashicorp/vault/api"
)

type VaultConfig struct {
	Address string
	Token   string
}

var (
	clientInstance *api.Client
	clientOnce     sync.Once
	clientMutex    sync.Mutex
)

// InitializeVaultClient initializes a singleton Vault client
func InitializeVaultClient(config *VaultConfig) (*api.Client, error) {
	var err error
	clientOnce.Do(func() {
		vaultConfig := api.DefaultConfig()
		vaultConfig.Address = config.Address

		clientInstance, err = api.NewClient(vaultConfig)
		if err != nil {
			return
		}

		clientInstance.SetToken(config.Token)
	})
	return clientInstance, err
}

// CreateTenantPolicy creates a policy for a specific tenant
func CreateTenantPolicy(client *api.Client, tenantid string) error {
	policyName := fmt.Sprintf("%s-policy", tenantid)
	policy := fmt.Sprintf(`
path "secret/data/tenants/%s/*" {
  capabilities = ["read"]
}
`, tenantid)

	err := client.Sys().PutPolicy(policyName, policy)
	if err != nil {
		return fmt.Errorf("failed to create policy: %v", err)
	}

	return nil
}

// CreateTenantToken creates a token for a specific tenant with the appropriate policy
func CreateTenantToken(client *api.Client, tenantid string) (string, error) {
	policyName := fmt.Sprintf("%s-policy", tenantid)
	tokenRequest := &api.TokenCreateRequest{
		Policies: []string{policyName},
		TTL:      "1h", // Short-lived token
	}

	secret, err := client.Auth().Token().Create(tokenRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %v", err)
	}

	return secret.Auth.ClientToken, nil
}

// RefreshToken refreshes the token if it is expired or about to expire
func RefreshToken(client *api.Client, config *VaultConfig) error {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Check token expiration
	secret, err := client.Auth().Token().LookupSelf()
	if err != nil {
		return fmt.Errorf("failed to lookup token: %v", err)
	}

	expireTime, ok := secret.Data["expire_time"].(string)
	if !ok {
		return fmt.Errorf("failed to parse token expiration time")
	}

	expireTimeParsed, err := time.Parse(time.RFC3339, expireTime)
	if err != nil {
		return fmt.Errorf("failed to parse token expiration time: %v", err)
	}

	// Refresh token if it is about to expire
	if time.Until(expireTimeParsed) < 10*time.Minute {
		newToken, err := CreateTenantToken(client, "tenantid") // Replace "tenantid" with actual tenant ID
		if err != nil {
			return fmt.Errorf("failed to refresh token: %v", err)
		}
		client.SetToken(newToken)
		config.Token = newToken
	}

	return nil
}

// RetrieveSecret retrieves a secret from Vault, refreshing the token if necessary
func RetrieveSecret(config *VaultConfig, secretPath string) (map[string]interface{}, error) {
	// Initialize Vault client
	client, err := InitializeVaultClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %v", err)
	}

	// Refresh token if necessary
	err = RefreshToken(client, config)
	if err != nil {
		return nil, err
	}

	// Read the secret from the specified path
	secret, err := client.Logical().Read(secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %v", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path: %s", secretPath)
	}

	return secret.Data, nil
}

// GetClientCredentials retrieves the credentials for a specific client, refreshing the token if necessary
func GetClientCredentials(config *VaultConfig, tenantid string) (map[string]interface{}, error) {
	secretPath := fmt.Sprintf("secret/data/tenants/%s/credentials", tenantid)

	// Retrieve secret with token refresh
	credentials, err := RetrieveSecret(config, secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve client credentials: %v", err)
	}

	return credentials, nil
}
