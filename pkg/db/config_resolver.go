package db

import (
	"context"
	"fmt"

	"getnoti.com/config"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/vault"
)

// DatabaseConfig represents tenant database configuration
type DatabaseConfig struct {
	TenantID     string
	Driver       string
	Host         string
	Port         int
	DatabaseName string
	Username     string
	Password     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int
}

// ConfigResolver interface for resolving tenant database configurations
type ConfigResolver interface {
	ResolveConfig(ctx context.Context, tenantID string) (*DatabaseConfig, error)
	CreateTenantConfig(ctx context.Context, tenantID string, config *DatabaseConfig) error
}

// configResolver implements ConfigResolver
type configResolver struct {
	config            *config.Config
	logger            logger.Logger
	credentialService CredentialService
}

// NewConfigResolver creates a new database config resolver
func NewConfigResolver(
	config *config.Config,
	logger logger.Logger,
	credentialService CredentialService,
) ConfigResolver {
	return &configResolver{
		config:            config,
		logger:            logger,
		credentialService: credentialService,
	}
}

// ResolveConfig resolves database configuration for a tenant
func (r *configResolver) ResolveConfig(ctx context.Context, tenantID string) (*DatabaseConfig, error) {
	r.logger.DebugContext(ctx, "Resolving database config for tenant",
		logger.String("tenant_id", tenantID))

	switch r.config.Credentials.StorageType {
	case "vault":
		return r.resolveFromVault(ctx, tenantID)
	case "database":
		return r.resolveFromDatabase(ctx, tenantID)
	case "auto":
		// Try vault first, fallback to database, then default
		if config, err := r.resolveFromVault(ctx, tenantID); err == nil {
			return config, nil
		}
		if config, err := r.resolveFromDatabase(ctx, tenantID); err == nil {
			return config, nil
		}
		return r.resolveDefault(ctx, tenantID)
	default:
		return r.resolveDefault(ctx, tenantID)
	}
}

// resolveFromVault resolves configuration from vault
func (r *configResolver) resolveFromVault(ctx context.Context, tenantID string) (*DatabaseConfig, error) {
	r.logger.DebugContext(ctx, "Resolving config from vault",
		logger.String("tenant_id", tenantID))

	// Get credentials from vault
	credentials, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "default")
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials from vault: %w", err)
	}

	return r.buildConfigFromMap(tenantID, credentials)
}

// resolveFromDatabase resolves configuration from database
func (r *configResolver) resolveFromDatabase(ctx context.Context, tenantID string) (*DatabaseConfig, error) {
	r.logger.DebugContext(ctx, "Resolving config from database",
		logger.String("tenant_id", tenantID))

	if r.credentialService == nil {
		return nil, fmt.Errorf("credential service not available")
	}

	credentials, err := r.credentialService.GetTenantDatabaseCredentials(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials from database: %w", err)
	}

	return r.buildConfigFromMap(tenantID, credentials)
}

// resolveDefault resolves default configuration
func (r *configResolver) resolveDefault(ctx context.Context, tenantID string) (*DatabaseConfig, error) {
	r.logger.DebugContext(ctx, "Using default config",
		logger.String("tenant_id", tenantID))

	return &DatabaseConfig{
		TenantID:     tenantID,
		Driver:       r.config.Database.Type,
		Host:         "localhost",
		Port:         5432,
		DatabaseName: fmt.Sprintf("tenant_%s", tenantID),
		Username:     "postgres",
		Password:     "password",
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  300,
	}, nil
}

// buildConfigFromMap builds DatabaseConfig from credentials map
func (r *configResolver) buildConfigFromMap(tenantID string, credentials map[string]interface{}) (*DatabaseConfig, error) {
	config := &DatabaseConfig{
		TenantID:     tenantID,
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  300,
	}

	// Extract required fields
	if driver, ok := credentials["type"].(string); ok {
		config.Driver = driver
	} else {
		return nil, fmt.Errorf("missing or invalid database type")
	}

	// If DSN is provided, we can use it directly
	if dsn, ok := credentials["dsn"].(string); ok {
		// For DSN-based configs, we still need the driver type
		config.DatabaseName = dsn // Store DSN in DatabaseName for now
		return config, nil
	}

	// Extract individual connection parameters
	if host, ok := credentials["host"].(string); ok {
		config.Host = host
	}

	if port, ok := credentials["port"]; ok {
		switch v := port.(type) {
		case int:
			config.Port = v
		case float64:
			config.Port = int(v)
		case string:
			// Try to parse string to int
			if p := parsePortString(v); p > 0 {
				config.Port = p
			}
		}
	}

	if dbName, ok := credentials["database"].(string); ok {
		config.DatabaseName = dbName
	}

	if username, ok := credentials["username"].(string); ok {
		config.Username = username
	}

	if password, ok := credentials["password"].(string); ok {
		config.Password = password
	}

	if sslMode, ok := credentials["ssl_mode"].(string); ok {
		config.SSLMode = sslMode
	} else {
		config.SSLMode = "disable" // Default
	}

	return config, nil
}

// CreateTenantConfig creates new database configuration for a tenant
func (r *configResolver) CreateTenantConfig(ctx context.Context, tenantID string, config *DatabaseConfig) error {
	r.logger.DebugContext(ctx, "Creating tenant config",
		logger.String("tenant_id", tenantID))

	// Convert config to credentials map
	credentials := map[string]interface{}{
		"type":     config.Driver,
		"host":     config.Host,
		"port":     fmt.Sprintf("%d", config.Port),
		"database": config.DatabaseName,
		"username": config.Username,
		"password": config.Password,
		"sslmode":  config.SSLMode,
	}

	// Store based on configured storage type
	switch r.config.Credentials.StorageType {
	case "vault":
		return vault.CreateCredential(tenantID, vault.DBCredential, "default", credentials)
	case "database":
		if r.credentialService == nil {
			return fmt.Errorf("credential service not available")
		}
		return r.credentialService.StoreTenantDatabaseCredentials(tenantID, credentials)
	default:
		return fmt.Errorf("storage type %s does not support creating tenant configs", r.config.Credentials.StorageType)
	}
}

// parsePortString parses port from string
func parsePortString(s string) int {
	switch s {
	case "5432":
		return 5432
	case "3306":
		return 3306
	case "1433":
		return 1433
	default:
		return 0
	}
}
