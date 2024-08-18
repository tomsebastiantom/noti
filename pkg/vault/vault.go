package vault

import (
	"fmt"
	"sync"

	// "time"

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
type TokenRequest struct {
	Type     string // "admin" or "tenant"
	TenantID string // Optional for admin tokens
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

//ToDo make it thread safe and make it clone client
//and also cache stuff maybe not as database client
//is cached but we need clear cache if updated

func Initialize(cfg *VaultConfig) error {
	var err error
	once.Do(func() {
		fmt.Println("Initializing Vault...")

		if cfg == nil {
			err = fmt.Errorf("vault configuration is nil")
			fmt.Println("Error: Vault configuration is nil")
			return
		}

		//fmt.Printf("Vault Address: %s, Token: %s\n", cfg.Address, cfg.Token)

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

		//Long Lived admin token used
		client.SetToken(config.Token)

		fmt.Println("Vault initialization completed successfully")
	})
	return err
}

func generateToken(req TokenRequest) (string, error) {
	var policies []string
	var ttl string

	switch req.Type {
	case "admin":
		policies = []string{"admin-policy"}
		ttl = "768h"
	case "tenant":
		if req.TenantID == "" {
			return "", fmt.Errorf("tenant ID is required for tenant token")
		}
		policyName := fmt.Sprintf("%s-policy", req.TenantID)
		policies = []string{policyName}
		ttl = "1h"
	default:
		return "", fmt.Errorf("unknown token type: %s", req.Type)
	}

	tokenRequest := &api.TokenCreateRequest{
		Policies:  policies,
		TTL:       ttl,
		Renewable: func(b bool) *bool { return &b }(true),
	}

	secret, err := client.Auth().Token().Create(tokenRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create %s token: %v", req.Type, err)
	}

	fmt.Printf("Created %s token: %s\n", req.Type, secret.Auth.ClientToken)
	return secret.Auth.ClientToken, nil
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

func updateTenantToken(tenantID string) error {
	newToken, err := generateToken(TokenRequest{Type: "tenant", TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("failed to update tenant token: %v", err)
	}

	client.SetToken(newToken)
	return nil
}

func CreateCredential(tenantID string, credType CredentialType, name string, data map[string]interface{}) error {
	//Setting admin token
	client.SetToken(config.Token)

	err := ensureTenantPolicy(tenantID)
	if err != nil {
		return err
	}

	// Ensure the KV secrets engine is mounted
	err = ensureKVEngineMounted(client, "secret/")
	if err != nil {
		return err
	}

	err = updateTenantToken(tenantID)
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
		return fmt.Errorf("failed to create credential: %v", err)
	}

	return nil
}

func GetClientCredentials(tenantID string, credType CredentialType, name string) (map[string]interface{}, error) {

	//Setting admin token
	client.SetToken(config.Token)

	err := updateTenantToken(tenantID)
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

    //Setting admin token
	client.SetToken(config.Token)

	err := updateTenantToken(tenantID)
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
