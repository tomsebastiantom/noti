package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
    App         AppConfig         `koanf:"app"`
    HTTP        HTTPConfig        `koanf:"http"`
    Logger      LoggerConfig      `koanf:"logger"`
    Database    DatabaseConfig    `koanf:"database"`
    Queue       QueueConfig       `koanf:"queue"`
    Vault       VaultConfig       `koanf:"vault"`
    Credentials CredentialsConfig `koanf:"credentials"`
    Env         string            `koanf:"env"`
}

type AppConfig struct {
    Name    string `koanf:"name"`
    Version string `koanf:"version"`
}

type HTTPConfig struct {
    Port string `koanf:"port"`
}

type LoggerConfig struct {
    Level string `koanf:"log_level"`
}

type DatabaseConfig struct {
    Type           string `koanf:"type"`
    DSN            string `koanf:"dsn"`
    MigrateOnStart bool   `koanf:"migrate_on_start"`
}

type QueueConfig struct {
    Enabled              bool          `koanf:"enabled"`
    URL                  string        `koanf:"url"`
    ReconnectInterval    time.Duration `koanf:"reconnect_interval"`
    MaxReconnectAttempts int           `koanf:"max_reconnect_attempts"`
    HeartbeatInterval    time.Duration `koanf:"heartbeat_interval"`
}

type VaultConfig struct {
    Enabled  bool   `koanf:"enabled"`
    Address  string `koanf:"address"`
    Provider string `koanf:"provider"` // "hashicorp", "aws", "azure", "gcp"
    // Token removed - comes from environment only
}

type CredentialsConfig struct {
    StorageType        string `koanf:"storage_type"`        // "vault", "database", "auto"
    EncryptionKeyEnv   string `koanf:"encryption_key_env"`  // Environment variable name for encryption key
    AllowCustomKeys    bool   `koanf:"allow_custom_keys"`   // Allow tenants to bring their own keys
    DefaultToDatabase  bool   `koanf:"default_to_database"` // Default to database if vault fails
}

var k = koanf.New(".")

func LoadConfig() (*Config, error) {
    // 1. Load defaults first
    if err := loadDefaults(); err != nil {
        return nil, fmt.Errorf("failed to load defaults: %w", err)
    }

    // 2. Load config file with discovery fallback
    if err := loadConfigWithDiscovery(); err != nil {
        logConfigWarning("Config file not loaded: %v", err)
    }

    // 3. Load environment variables (highest priority)
    if err := loadEnvironmentVariables(); err != nil {
        return nil, fmt.Errorf("failed to load environment variables: %w", err)
    }

    return unmarshalAndValidate()
}

func loadConfigWithDiscovery() error {
    // Priority order for config loading:
    
    // 1. Explicit path (highest priority)
    if configPath := getExplicitConfigPath(); configPath != "" {
        return loadConfigFile(configPath)
    }
    
    // 2. Auto-discovery (development convenience)
    if configPath := discoverConfigFile(); configPath != "" {
        logConfigInfo("Auto-discovered config: %s", configPath)
        return loadConfigFile(configPath)
    }
    
    return fmt.Errorf("no config file found")
}

func getExplicitConfigPath() string {
    // User-provided paths (like Kubernetes)
    if configPath := os.Getenv("NOTI_CONFIG"); configPath != "" {
        return configPath
    }
    
    if configPath := os.Getenv("NOTI_CONFIG_PATH"); configPath != "" {
        return configPath
    }
    
    return ""
}

func discoverConfigFile() string {
    // Auto-discovery candidates (like Docker/Git)
    candidates := []string{
        "./config.yaml",         // Current directory
        "./config/config.yaml",  // Config subdirectory
        "./noti.yaml",          // Alternative name
    }
    
    for _, candidate := range candidates {
        if _, err := os.Stat(candidate); err == nil {
            // Convert to absolute path for clarity
            if absPath, err := filepath.Abs(candidate); err == nil {
                return absPath
            }
            return candidate
        }
    }
    
    return ""
}

func loadDefaults() error {
    defaults := map[string]interface{}{
        // App defaults
        "app.name":                        "noti",
        "app.version":                     "1.0.0",
        
        // HTTP defaults
        "http.port":                       "8072",
        
        // Logger defaults
        "logger.log_level":                "debug",
        
        // Database defaults (SQLite for easy development)
        "database.type":                   "sqlite",
        "database.dsn":                    "./data/noti.db",
        "database.migrate_on_start":       true,
        
        // Queue defaults (disabled by default for development)
        "queue.enabled":                   false,
        "queue.url":                       "",
        "queue.reconnect_interval":        "5s",
        "queue.max_reconnect_attempts":    3,
        "queue.heartbeat_interval":        "30s",
        
        // Vault defaults (disabled by default for development)
        "vault.enabled":                   false,
        "vault.address":                   "",
        "vault.provider":                  "hashicorp",
        
        // Credentials defaults
        "credentials.storage_type":        "auto",
        "credentials.encryption_key_env":  "NOTI_ENCRYPTION_KEY",
        "credentials.allow_custom_keys":   true,
        "credentials.default_to_database": true,
        
        // Environment
        "env":                            "development",
    }

    return k.Load(confmap.Provider(defaults, "."), nil)
}

func loadConfigFile(configPath string) error {
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return fmt.Errorf("config file '%s' does not exist", configPath)
    }

    if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
        return fmt.Errorf("failed to load config file '%s': %w", configPath, err)
    }
    
    logConfigInfo("Loaded config from: %s", configPath)
    return nil
}

func loadEnvironmentVariables() error {
    return k.Load(env.Provider("NOTI_", ".", func(s string) string {
        key := strings.ToLower(strings.TrimPrefix(s, "NOTI_"))
        return strings.ReplaceAll(key, "_", ".")
    }), nil)
}

func unmarshalAndValidate() (*Config, error) {
    var cfg Config
    if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    if err := validateConfig(&cfg); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    logConfigInfo("Configuration loaded successfully for environment: %s", cfg.Env)
    return &cfg, nil
}

func validateConfig(cfg *Config) error {
    // Basic validation
    if cfg.App.Name == "" {
        return fmt.Errorf("app.name is required")
    }
    if cfg.HTTP.Port == "" {
        return fmt.Errorf("http.port is required")
    }
    
    // Validate enabled services have required config
    if cfg.Vault.Enabled {
        if cfg.Vault.Address == "" {
            return fmt.Errorf("vault.address is required when vault is enabled")
        }
        if os.Getenv("NOTI_VAULT_TOKEN") == "" {
            return fmt.Errorf("NOTI_VAULT_TOKEN environment variable is required when vault is enabled")
        }
    }
    
    if cfg.Queue.Enabled {
        queueURL := cfg.Queue.URL
        if envURL := os.Getenv("NOTI_QUEUE_URL"); envURL != "" {
            queueURL = envURL
        }
        if queueURL == "" {
            return fmt.Errorf("queue URL is required when queue is enabled (set NOTI_QUEUE_URL or queue.url)")
        }
    }
    
    // Validate encryption key for database credential storage
    if cfg.Credentials.StorageType == "database" || cfg.Credentials.StorageType == "auto" {
        if os.Getenv(cfg.Credentials.EncryptionKeyEnv) == "" {
            return fmt.Errorf("encryption key is required for database credential storage (set %s)", cfg.Credentials.EncryptionKeyEnv)
        }
    }
    
    return nil
}

func logConfigInfo(format string, args ...interface{}) {
    if os.Getenv("NOTI_CONFIG_DEBUG") == "true" {
        fmt.Printf("[CONFIG] "+format+"\n", args...)
    }
}

func logConfigWarning(format string, args ...interface{}) {
    if os.Getenv("NOTI_ENV") != "production" {
        fmt.Printf("[CONFIG WARN] "+format+"\n", args...)
    }
}

// Helper methods to get secrets from environment variables
func (c *Config) GetVaultToken() string {
    return os.Getenv("NOTI_VAULT_TOKEN")
}

func (c *Config) GetDatabaseDSN() string {
    if dsn := os.Getenv("NOTI_DATABASE_DSN"); dsn != "" {
        return dsn
    }
    return c.Database.DSN // Fallback to config
}

func (c *Config) GetQueueURL() string {
    if url := os.Getenv("NOTI_QUEUE_URL"); url != "" {
        return url
    }
    return c.Queue.URL // Fallback to config
}

func (c *Config) GetEncryptionKey() string {
    return os.Getenv(c.Credentials.EncryptionKeyEnv)
}