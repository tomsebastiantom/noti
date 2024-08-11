package db

import (
	"fmt"

	"sync"

	"getnoti.com/config"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/vault"

	_ "github.com/denisenkom/go-mssqldb" // Microsoft SQL Server driver
	_ "github.com/go-sql-driver/mysql"   // MySQL driver
	_ "github.com/lib/pq"                // PostgreSQL driver
)

// Manager manages database connections and caches them.

type Manager struct {
	cache  *cache.GenericCache
	mutex  sync.Mutex
	config *config.Config
	logger *logger.Logger
}

// NewManager creates a new Manager instance.
func NewManager(cache *cache.GenericCache, vaultConfig *vault.VaultConfig, config *config.Config, logger *logger.Logger) *Manager {
	return &Manager{
		cache:  cache,
		logger: logger,
		config: config,
	}
}

// GetDatabaseConnection retrieves a database connection for the given tenant ID.
func (m *Manager) GetDatabaseConnection(tenantID string) (Database, error) {
	dbConn, found := m.cache.Get(tenantID)
	if !found {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		dbConn, found = m.cache.Get(tenantID)
		if !found {
			var err error
			dbConn, err = m.createDatabaseConnectionFromVault(tenantID)
			if err != nil {
				return nil, err
			}
			m.cache.Set(tenantID, dbConn, 1)
		}
	}

	return dbConn.(Database), nil
}

// GetDatabaseConnectionWithConfig retrieves a database connection using the provided configuration.
func (m *Manager) GetDatabaseConnectionWithConfig(tenantID string, dbConfig map[string]interface{}) (Database, error) {
	if len(dbConfig) == 0 {
		return m.GetDatabaseConnection(tenantID)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	dbConn, found := m.cache.Get(tenantID)
	if !found {
		var err error
		dbConn, err = m.createDatabaseConnectionFromConfig(dbConfig)
		if err != nil {
			return nil, err
		}
		m.cache.Set(tenantID, dbConn, 1)
	}

	return dbConn.(Database), nil
}

func (m *Manager) createDatabaseConnectionFromVault(tenantID string) (Database, error) {
	m.logger.Debug(fmt.Sprintf("Creating new database connection for tenant: %s from vault", tenantID))
	credentials, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get database credentials for tenant %s: %v", tenantID, err)
	}

	dbConfig := make(map[string]interface{})
	if dsn, ok := credentials["dsn"].(string); ok {
		dbConfig["type"] = credentials["type"]
		dbConfig["dsn"] = dsn
	} else {
		dbConfig["type"] = credentials["type"]
		dbConfig["host"] = credentials["host"]
		dbConfig["port"] = credentials["port"]
		dbConfig["username"] = credentials["username"]
		dbConfig["password"] = credentials["password"]
		dbConfig["database"] = credentials["database"]
	}

	return m.createDatabaseConnectionFromConfig(dbConfig)
}

func (m *Manager) createDatabaseConnectionFromConfig(dbConfig map[string]interface{}) (Database, error) {
	m.logger.Debug("Creating new database connection from provided configuration")
	dbConn, err := NewDatabaseFactory(dbConfig, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %v", err)
	}
	return dbConn, nil
}

func (m *Manager) CreateNewTenantDatabase(tenantID string) (Database, map[string]interface{}, error) {

	dbConfig := map[string]interface{}{
		"type": m.config.Database.Type,
		"dsn":  m.config.Database.DSN,
	}

	dbConn, newDbConfig, err := createTenantDatabase(tenantID, dbConfig, m.logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new tenant database for %s: %v", tenantID, err)
	}

	m.cache.Set(tenantID, dbConn, 1)

	return dbConn, newDbConfig, nil
}
