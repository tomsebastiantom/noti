package db

import (
    "fmt"
    "sync"
    
    "getnoti.com/config"
    "getnoti.com/pkg/cache"
    "getnoti.com/pkg/logger"
    
    _ "github.com/denisenkom/go-mssqldb" // Microsoft SQL Server driver
    _ "github.com/go-sql-driver/mysql"   // MySQL driver
    _ "github.com/lib/pq"                // PostgreSQL driver
)

// Manager manages database connections and caches them.
type Manager struct {
    cache             *cache.GenericCache
    mutex             sync.Mutex
    config            *config.Config
    logger            logger.Logger
    mainDB            Database
    credentialService CredentialService
}

// CredentialService interface for getting tenant database credentials
type CredentialService interface {
    GetTenantDatabaseCredentials(tenantID string) (map[string]interface{}, error)
    StoreTenantDatabaseCredentials(tenantID string, credentials map[string]interface{}) error
}

// NewManager creates a new Manager instance.
func NewManager(cache *cache.GenericCache, config *config.Config, logger logger.Logger, mainDB Database) *Manager {
    return &Manager{
        cache:  cache,
        logger: logger,
        config: config,
        mainDB: mainDB,
    }
}

// SetCredentialService sets the credential service (dependency injection)
func (m *Manager) SetCredentialService(service CredentialService) {
    m.credentialService = service
}

// GetDatabaseConnection retrieves a database connection for the given tenant ID.
func (m *Manager) GetDatabaseConnection(tenantID string) (Database, error) {
    // Check cache first
    if dbConn, found := m.cache.Get(tenantID); found {
        return dbConn.(Database), nil
    }

    m.mutex.Lock()
    defer m.mutex.Unlock()

    // Double-check cache after acquiring lock
    if dbConn, found := m.cache.Get(tenantID); found {
        return dbConn.(Database), nil
    }

    // Create new connection
    dbConn, err := m.createDatabaseConnection(tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to create database connection for tenant %s: %w", tenantID, err)
    }

    // Cache the connection
    m.cache.Set(tenantID, dbConn, 1)
    return dbConn, nil
}

// GetDatabaseConnectionWithConfig retrieves a database connection using the provided configuration.
func (m *Manager) GetDatabaseConnectionWithConfig(tenantID string, dbConfig map[string]interface{}) (Database, error) {
    if len(dbConfig) == 0 {
        return m.GetDatabaseConnection(tenantID)
    }

    cacheKey := fmt.Sprintf("%s:custom", tenantID)
    
    m.mutex.Lock()
    defer m.mutex.Unlock()

    if dbConn, found := m.cache.Get(cacheKey); found {
        return dbConn.(Database), nil
    }

    dbConn, err := m.createDatabaseConnectionFromConfig(dbConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create database connection with config for tenant %s: %w", tenantID, err)
    }

    m.cache.Set(cacheKey, dbConn, 1)
    return dbConn, nil
}

func (m *Manager) createDatabaseConnection(tenantID string) (Database, error) {
    // Try to get tenant-specific database credentials
    if m.credentialService != nil {
        if dbConfig, err := m.credentialService.GetTenantDatabaseCredentials(tenantID); err == nil {
            m.logger.Debug(fmt.Sprintf("Creating database connection for tenant %s from stored credentials", tenantID))
            return m.createDatabaseConnectionFromConfig(dbConfig)
        } else {
            m.logger.Debug(fmt.Sprintf("No stored credentials found for tenant %s, using main database: %v", tenantID, err))
        }
    }

    // Fallback to main database (shared tenant approach)
    m.logger.Debug(fmt.Sprintf("Creating database connection for tenant %s using main database", tenantID))
    return m.mainDB, nil
}

func (m *Manager) createDatabaseConnectionFromConfig(dbConfig map[string]interface{}) (Database, error) {
    m.logger.Debug("Creating new database connection from provided configuration")
    dbConn, err := NewDatabaseFactory(dbConfig, m.logger)
    if err != nil {
        return nil, fmt.Errorf("failed to create database connection: %w", err)
    }
    return dbConn, nil
}

func (m *Manager) CreateNewTenantDatabase(tenantID string) (Database, map[string]interface{}, error) {
    m.logger.Info(fmt.Sprintf("Creating new dedicated database for tenant: %s", tenantID))
    
    // Create new database for tenant based on main config
    dbConfig := map[string]interface{}{
        "type": m.config.Database.Type,
        "dsn":  m.config.Database.DSN,
    }

    dbConn, newDbConfig, err := createTenantDatabase(tenantID, dbConfig, m.logger)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to create new tenant database for %s: %w", tenantID, err)
    }

    // Store the credentials for future use
    if m.credentialService != nil {
        if err := m.credentialService.StoreTenantDatabaseCredentials(tenantID, newDbConfig); err != nil {
            m.logger.Warn(fmt.Sprintf("Failed to store database credentials for tenant %s: %v", tenantID, err))
        }
    }

    // Cache the connection
    m.cache.Set(tenantID, dbConn, 1)
    return dbConn, newDbConfig, nil
}

// InvalidateCache removes a tenant's database connection from cache
func (m *Manager) InvalidateCache(tenantID string) {
    m.cache.Delete(tenantID)
    m.cache.Delete(fmt.Sprintf("%s:custom", tenantID))
}

// Close closes all cached database connections
func (m *Manager) Close() error {
    // Note: This would require extending the cache to track all connections
    // For now, connections will be closed when they're garbage collected
    return nil
}