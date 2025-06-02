package config

import (
    "fmt"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/confmap"
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/v2"
    "os"
    "path/filepath"
    "strings"
    "time"
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
    Type string `koanf:"type"`
    DSN  string `koanf:"dsn"`
}

type QueueConfig struct {
    URL                  string        `koanf:"url"`
    ReconnectInterval    time.Duration `koanf:"reconnect_interval"`
    MaxReconnectAttempts int           `koanf:"max_reconnect_attempts"`
    HeartbeatInterval    time.Duration `koanf:"heartbeat_interval"`
}

type VaultConfig struct {
    Address  string `koanf:"address"`
    Token    string `koanf:"token"`
    Provider string `koanf:"provider"` // "hashicorp", "aws", "azure", "gcp"
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
        "app.name":                        "noti",
        "app.version":                     "1.0.0",
        "http.port":                       "8072",
        "logger.log_level":                "debug",
        "database.type":                   "sqlite",
        "database.dsn":                    "./data/noti.db",
        "queue.url":                       "",
        "queue.reconnect_interval":        "5s",
        "queue.max_reconnect_attempts":    3,
        "queue.heartbeat_interval":        "30s",
        "vault.address":                   "",
        "vault.token":                     "",
        "vault.provider":                  "hashicorp",
        "credentials.storage_type":        "auto",
        "credentials.encryption_key_env":  "NOTI_ENCRYPTION_KEY",
        "credentials.allow_custom_keys":   true,
        "credentials.default_to_database": true,
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
    if cfg.App.Name == "" {
        return fmt.Errorf("app.name is required")
    }
    if cfg.HTTP.Port == "" {
        return fmt.Errorf("http.port is required")
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