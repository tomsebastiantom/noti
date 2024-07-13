package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/vault"
	_ "github.com/go-sql-driver/mysql"    // MySQL driver
	_ "github.com/lib/pq"                 // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3"       // SQLite driver
	_ "github.com/denisenkom/go-mssqldb"  // Microsoft SQL Server driver
	_ "github.com/godror/godror"          // Oracle driver
)

// Database interface defines the methods for database operations.
type Database interface {
    Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// SQLDatabase wraps the sql.DB to implement the Database interface.
type SQLDatabase struct {
    *sql.DB
}

func (db *SQLDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    return db.DB.QueryContext(ctx, query, args...)
}

func (db *SQLDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    return db.DB.ExecContext(ctx, query, args...)
}

func (db *SQLDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
    return db.DB.QueryRowContext(ctx, query, args...)
}

// NewDatabaseFactory initializes a new database connection based on the provided configuration.
func NewDatabaseFactory(cfg *DatabaseConfig) (Database, error) {
    db, err := sql.Open(cfg.Type, cfg.DSN)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %v", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %v", err)
    }

    return &SQLDatabase{db}, nil
}

// DatabaseConfig holds the configuration for the database connection.
type DatabaseConfig struct {
    Type string
    DSN  string
}

// Manager manages database connections and caches them.
type Manager struct {
    cache       *cache.GenericCache
    vaultConfig *vault.VaultConfig
    mutex       sync.Mutex
}

// NewManager creates a new Manager instance.
func NewManager(cache *cache.GenericCache, vaultConfig *vault.VaultConfig) *Manager {
    return &Manager{
        cache:       cache,
        vaultConfig: vaultConfig,
    }
}

// GetDatabaseConnection retrieves a database connection for the given tenant ID.
func (m *Manager) GetDatabaseConnection(tenantID string) (Database, error) {
    // Check if the database connection is cached
    dbConn, found := m.cache.Get(tenantID)
    if !found {
        m.mutex.Lock()
        defer m.mutex.Unlock()

        // Double-check if the database connection is still not cached
        dbConn, found = m.cache.Get(tenantID)
        if !found {
            // Retrieve database credentials from Vault
            credentials, err := vault.GetClientCredentials(m.vaultConfig, tenantID)
            if err != nil {
                return nil, fmt.Errorf("failed to retrieve database credentials: %v", err)
            }

            // Create a new database connection
            dbConfig := &DatabaseConfig{
                Type: credentials["type"].(string),
                DSN:  credentials["dsn"].(string),
            }
            dbConn, err = NewDatabaseFactory(dbConfig)
            if err != nil {
                return nil, fmt.Errorf("failed to create database connection: %v", err)
            }

            // Cache the database connection
            m.cache.Set(tenantID, dbConn, 1)
        }
    }

    return dbConn.(Database), nil
}
