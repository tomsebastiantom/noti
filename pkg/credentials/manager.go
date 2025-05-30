package credentials

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "sync"

    "getnoti.com/config"
    "getnoti.com/pkg/db"
    "getnoti.com/pkg/errors"
    log "getnoti.com/pkg/logger"
    "getnoti.com/pkg/vault"
    "golang.org/x/crypto/pbkdf2"
)

type Manager struct {
    config            *config.Config
    logger            log.Logger
    mainDB            db.Database
    vaultEnabled      bool
    vaultInitialized  bool
    encryptionKey     []byte
    tenantKeys        map[string]cipher.AEAD
    keysMutex         sync.RWMutex
    operationMutex    sync.Mutex
}

// StorageType represents where credentials are stored
type StorageType string

const (
    StorageVault    StorageType = "vault"
    StorageDatabase StorageType = "database"
    StorageAuto     StorageType = "auto"
)

// CredentialType from your existing vault package
type CredentialType = vault.CredentialType

// Use existing types from vault package
const (
    DBCredential      = vault.DBCredential
    GenericCredential = vault.GenericCredential
)

// TenantEncryptionConfig for custom tenant keys
type TenantEncryptionConfig struct {
    UseCustomKey      bool   `json:"use_custom_key"`
    CustomKeyHash     string `json:"custom_key_hash,omitempty"`
    KeyDerivationSalt string `json:"key_derivation_salt,omitempty"`
    EncryptionVersion int    `json:"encryption_version"`
}

func NewManager(config *config.Config, logger log.Logger, mainDB db.Database) (*Manager, error) {
    ctx := context.Background()
    
    m := &Manager{
        config:     config,
        logger:     logger,
        mainDB:     mainDB,
        tenantKeys: make(map[string]cipher.AEAD),
    }

    // Initialize encryption key for database storage
    if err := m.initEncryptionKey(); err != nil {
        return nil, errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("credential_manager_init").
            WithCause(err).
            WithMessage("Failed to initialize encryption key").
            Build()
    }

    // Initialize vault if configured
    if config.Vault.Address != "" && config.Vault.Token != "" {
        vaultConfig := &vault.VaultConfig{
            Address:  config.Vault.Address,
            Token:    config.Vault.Token,
            Provider: config.Vault.Provider,
        }
        
        if err := vault.Initialize(vaultConfig); err != nil {
            if config.Credentials.DefaultToDatabase {
				m.logger.Warn("Vault initialization failed, falling back to database storage",
				log.Err(err))
                m.vaultEnabled = false
            } else {
                return nil, errors.New(errors.ErrCodeInternal).
                    WithContext(ctx).
                    WithOperation("vault_initialization").
                    WithCause(err).
                    WithMessage("Failed to initialize vault and database fallback disabled").
                    Build()
            }
        } else {
            m.vaultEnabled = true
            m.vaultInitialized = true
            m.logger.Info("Vault initialized successfully",
                log.String("provider", vault.GetProviderType()))
        }
    } else {
        m.logger.Info("No vault configuration found, using database storage")
        m.vaultEnabled = false
    }

    m.logger.InfoContext(ctx, "Credential manager initialized successfully",
        log.Bool("vault_enabled", m.vaultEnabled),
        log.String("storage_type", config.Credentials.StorageType),
        log.String("vault_provider", vault.GetProviderType()))

    return m, nil
}


func (m *Manager) GetTenantDatabaseCredentials(tenantID string) (map[string]interface{}, error) {
    return m.GetCredentials(tenantID, DBCredential, "default")
}


func (m *Manager) StoreTenantDatabaseCredentials(tenantID string, credentials map[string]interface{}) error {
    return m.StoreCredentials(tenantID, DBCredential, "default", credentials)
}


func (m *Manager) GetCredentials(tenantID string, credType CredentialType, name string) (map[string]interface{}, error) {
    ctx := context.Background()
    
    // Determine storage type
    storageType := m.determineStorageType(tenantID, credType)
    
    switch storageType {
    case StorageVault:
        if m.vaultEnabled && m.vaultInitialized {
            credentials, err := vault.GetClientCredentials(tenantID, credType, name)
            if err != nil {
                if m.config.Credentials.DefaultToDatabase {
					m.logger.Warn("Vault read failed, trying database fallback",
					log.String("tenant_id", tenantID),
					log.String("cred_type", string(credType)),
					log.String("name", name),
					log.Err(err)) 
                    return m.getFromDatabase(ctx, tenantID, credType, name)
                }
                return nil, errors.New(errors.ErrCodeNotFound).
                    WithContext(ctx).
                    WithOperation("get_credentials_vault").
                    WithCause(err).
                    WithMessage("Failed to get credentials from vault").
                    WithDetails(map[string]interface{}{
                        "tenant_id": tenantID,
                        "cred_type": string(credType),
                        "name":      name,
                    }).
                    Build()
            }
            return credentials, nil
        }
        fallthrough // If vault not available, try database
        
    case StorageDatabase:
        return m.getFromDatabase(ctx, tenantID, credType, name)
        
    default:
        return nil, errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("get_credentials").
            WithMessage("Invalid storage type").
            WithDetails(map[string]interface{}{
                "storage_type": string(storageType),
                "tenant_id":    tenantID,
            }).
            Build()
    }
}

// StoreCredentials stores credentials using vault or database fallback
func (m *Manager) StoreCredentials(tenantID string, credType CredentialType, name string, credentials map[string]interface{}) error {
    ctx := context.Background()
    
    // Determine storage type
    storageType := m.determineStorageType(tenantID, credType)
    
    switch storageType {
    case StorageVault:
        if m.vaultEnabled && m.vaultInitialized {
            err := vault.CreateCredential(tenantID, credType, name, credentials)
            if err != nil {
                if m.config.Credentials.DefaultToDatabase {
                    m.logger.Warn("Vault write failed, falling back to database",
                        log.String("tenant_id", tenantID),
                        log.String("cred_type", string(credType)),
                        log.String("name", name),
                        log.Err(err))
                    return m.storeInDatabase(ctx, tenantID, credType, name, credentials, StorageDatabase)
                }
                return errors.New(errors.ErrCodeInternal).
                    WithContext(ctx).
                    WithOperation("store_credentials_vault").
                    WithCause(err).
                    WithMessage("Failed to store credentials in vault").
                    WithDetails(map[string]interface{}{
                        "tenant_id": tenantID,
                        "cred_type": string(credType),
                        "name":      name,
                    }).
                    Build()
            }
            return nil
        }
        fallthrough // If vault not available, use database
        
    case StorageDatabase:
        return m.storeInDatabase(ctx, tenantID, credType, name, credentials, StorageDatabase)
        
    default:
        return errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("store_credentials").
            WithMessage("Invalid storage type").
            WithDetails(map[string]interface{}{
                "storage_type": string(storageType),
                "tenant_id":    tenantID,
            }).
            Build()
    }
}

// UpdateCredentials updates existing credentials
func (m *Manager) UpdateCredentials(tenantID string, credType CredentialType, name string, credentials map[string]interface{}) error {
    ctx := context.Background()
    
    storageType := m.determineStorageType(tenantID, credType)
    
    switch storageType {
    case StorageVault:
        if m.vaultEnabled && m.vaultInitialized {
            err := vault.UpdateCredential(tenantID, credType, name, credentials)
            if err != nil {
                if m.config.Credentials.DefaultToDatabase {
                    return m.storeInDatabase(ctx, tenantID, credType, name, credentials, StorageDatabase)
                }
                return errors.New(errors.ErrCodeInternal).
                    WithContext(ctx).
                    WithOperation("update_credentials_vault").
                    WithCause(err).
                    WithMessage("Failed to update credentials in vault").
                    Build()
            }
            return nil
        }
        fallthrough
        
    case StorageDatabase:
        return m.storeInDatabase(ctx, tenantID, credType, name, credentials, StorageDatabase)
        
    default:
        return errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("update_credentials").
            WithMessage("Invalid storage type").
            Build()
    }
}

// Helper methods remain the same...
func (m *Manager) determineStorageType(tenantID string, credType CredentialType) StorageType {
    switch m.config.Credentials.StorageType {
    case string(StorageVault):
        return StorageVault
    case string(StorageDatabase):
        return StorageDatabase
    case string(StorageAuto):
        if m.vaultEnabled && m.vaultInitialized {
            return StorageVault
        }
        return StorageDatabase
    default:
        return StorageDatabase
    }
}

// All other helper methods remain the same...
func (m *Manager) getFromDatabase(ctx context.Context, tenantID string, credType CredentialType, name string) (map[string]interface{}, error) {
    query := `
    SELECT encrypted_data 
    FROM tenant_credentials 
    WHERE tenant_id = ? AND credential_type = ? AND name = ?`

    var encryptedData string
    err := m.mainDB.QueryRow(ctx, query, tenantID, string(credType), name).Scan(&encryptedData)
    if err != nil {
        return nil, errors.New(errors.ErrCodeNotFound).
            WithContext(ctx).
            WithOperation("get_from_database").
            WithCause(err).
            WithMessage("Credentials not found").
            WithDetails(map[string]interface{}{
                "tenant_id": tenantID,
                "cred_type": string(credType),
                "name":      name,
            }).
            Build()
    }

    return m.decryptForTenant(ctx, tenantID, encryptedData)
}

func (m *Manager) storeInDatabase(ctx context.Context, tenantID string, credType CredentialType, name string, credentials map[string]interface{}, storageType StorageType) error {
    encryptedData, err := m.encryptForTenant(ctx, tenantID, credentials)
    if err != nil {
        return errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("store_in_database").
            WithCause(err).
            WithMessage("Failed to encrypt credentials").
            Build()
    }

    query := `
    INSERT OR REPLACE INTO tenant_credentials 
    (tenant_id, credential_type, name, encrypted_data, storage_type, updated_at) 
    VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

    _, err = m.mainDB.Exec(ctx, query, tenantID, string(credType), name, encryptedData, string(storageType))
    if err != nil {
        return errors.DatabaseError(ctx, "store_credentials", err)
    }

    return nil
}

func (m *Manager) getTenantEncryptionKey(ctx context.Context, tenantID string) (cipher.AEAD, error) {
    m.keysMutex.RLock()
    if key, exists := m.tenantKeys[tenantID]; exists {
        m.keysMutex.RUnlock()
        return key, nil
    }
    m.keysMutex.RUnlock()

    m.keysMutex.Lock()
    defer m.keysMutex.Unlock()

    // Double-check after acquiring write lock
    if key, exists := m.tenantKeys[tenantID]; exists {
        return key, nil
    }

    config, err := m.getTenantEncryptionConfig(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    var encryptionKey []byte
    if config.UseCustomKey {
        return nil, errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("get_tenant_encryption_key").
            WithMessage("Custom encryption key not available in cache").
            WithDetails(map[string]interface{}{
                "tenant_id": tenantID,
            }).
            Build()
    } else {
        encryptionKey = m.deriveKeyFromMaster(tenantID)
    }

    aead, err := m.createAEAD(encryptionKey)
    if err != nil {
        return nil, errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("get_tenant_encryption_key").
            WithCause(err).
            WithMessage("Failed to create AEAD cipher").
            Build()
    }

    m.tenantKeys[tenantID] = aead
    return aead, nil
}

func (m *Manager) getTenantEncryptionConfig(ctx context.Context, tenantID string) (*TenantEncryptionConfig, error) {
    query := `SELECT use_custom_key, custom_key_hash, key_derivation_salt, encryption_version 
              FROM tenant_encryption_config WHERE tenant_id = ?`
    
    var config TenantEncryptionConfig
    err := m.mainDB.QueryRow(ctx, query, tenantID).Scan(
        &config.UseCustomKey, 
        &config.CustomKeyHash, 
        &config.KeyDerivationSalt, 
        &config.EncryptionVersion)
    
    if err != nil {
        // No custom config found, use default
        return &TenantEncryptionConfig{
            UseCustomKey:      false,
            EncryptionVersion: 1,
        }, nil
    }
    
    return &config, nil
}

func (m *Manager) encryptForTenant(ctx context.Context, tenantID string, data map[string]interface{}) (string, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return "", err
    }

    aead, err := m.getTenantEncryptionKey(ctx, tenantID)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, aead.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    encryptedData := aead.Seal(nonce, nonce, jsonData, nil)
    return base64.StdEncoding.EncodeToString(encryptedData), nil
}

func (m *Manager) decryptForTenant(ctx context.Context, tenantID string, encryptedString string) (map[string]interface{}, error) {
    encryptedData, err := base64.StdEncoding.DecodeString(encryptedString)
    if err != nil {
        return nil, err
    }

    aead, err := m.getTenantEncryptionKey(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    nonceSize := aead.NonceSize()
    if len(encryptedData) < nonceSize {
        return nil, fmt.Errorf("encrypted data too short")
    }

    nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
    jsonData, err := aead.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    var data map[string]interface{}
    err = json.Unmarshal(jsonData, &data)
    if err != nil {
        return nil, err
    }

    return data, nil
}

func (m *Manager) deriveKeyFromMaster(tenantID string) []byte {
    return pbkdf2.Key([]byte(tenantID), m.encryptionKey, 100000, 32, sha256.New)
}

func (m *Manager) createAEAD(key []byte) (cipher.AEAD, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    return cipher.NewGCM(block)
}

func (m *Manager) initEncryptionKey() error {
    keyEnvVar := m.config.Credentials.EncryptionKeyEnv
    if keyEnvVar == "" {
        keyEnvVar = "NOTI_ENCRYPTION_KEY"
    }
    
    key := os.Getenv(keyEnvVar)
    if key == "" {
        // Generate a random key for development
        keyBytes := make([]byte, 32)
        if _, err := rand.Read(keyBytes); err != nil {
            return err
        }
        key = base64.StdEncoding.EncodeToString(keyBytes)
        m.logger.Warn("Generated temporary encryption key. Set environment variable for production",
            log.String("env_var", keyEnvVar))
    }

    keyBytes, err := base64.StdEncoding.DecodeString(key)
    if err != nil {
        return fmt.Errorf("invalid encryption key format: %w", err)
    }

    if len(keyBytes) != 32 {
        return fmt.Errorf("encryption key must be 32 bytes, got %d", len(keyBytes))
    }

    m.encryptionKey = keyBytes
    return nil
}

// IsHealthy checks if the credential manager is healthy
func (m *Manager) IsHealthy() bool {
    ctx := context.Background()
    return m.mainDB.Ping(ctx) == nil
}

// GetStorageInfo returns information about the storage backend
func (m *Manager) GetStorageInfo() map[string]interface{} {
    return map[string]interface{}{
        "vault_enabled":       m.vaultEnabled,
        "vault_initialized":   m.vaultInitialized,
        "vault_provider":      vault.GetProviderType(),
        "storage_type":        m.config.Credentials.StorageType,
        "allow_custom_keys":   m.config.Credentials.AllowCustomKeys,
        "default_to_database": m.config.Credentials.DefaultToDatabase,
    }
}