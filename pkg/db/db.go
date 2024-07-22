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
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Prepare(ctx context.Context, query string) (*sql.Stmt, error)
	Close() error
	Ping(ctx context.Context) error
}

// Transaction interface defines methods for database transactions.
type Transaction interface {
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Commit() error
	Rollback() error
}

// SQLDatabase wraps the sql.DB to implement the Database interface.
type SQLDatabase struct {
	*sql.DB
}

// Implement the Database interface methods for SQLDatabase
func (db *SQLDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *SQLDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *SQLDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *SQLDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &SQLTransaction{tx}, nil
}

func (db *SQLDatabase) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return db.DB.PrepareContext(ctx, query)
}

func (db *SQLDatabase) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

// SQLTransaction wraps sql.Tx to implement the Transaction interface
type SQLTransaction struct {
	*sql.Tx
}

// Implement the Transaction interface methods for SQLTransaction
func (tx *SQLTransaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return tx.Tx.ExecContext(ctx, query, args...)
}

func (tx *SQLTransaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return tx.Tx.QueryContext(ctx, query, args...)
}

func (tx *SQLTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return tx.Tx.QueryRowContext(ctx, query, args...)
}

// NewDatabaseFactory initializes a new database connection based on the provided configuration.
func NewDatabaseFactory(credentials map[string]interface{}) (Database, error) {
    var dbConfig map[string]string

    if dsn, ok := credentials["dsn"].(string); ok {
        dbConfig = map[string]string{
            "type": credentials["type"].(string),
            "dsn":  dsn,
        }
    } else {
        dbConfig = map[string]string{
            "type":     credentials["type"].(string),
            "host":     credentials["host"].(string),
            "port":     credentials["port"].(string),
            "username": credentials["username"].(string),
            "password": credentials["password"].(string),
            "database": credentials["database"].(string),
        }
    }

    return createDatabaseConnection(dbConfig)
}

func createDatabaseConnection(dbConfig map[string]string) (Database, error) {
    dbType, ok := dbConfig["type"]
    if !ok {
        return nil, fmt.Errorf("database type is required")
    }

    var dsn string
    if dsn, ok = dbConfig["dsn"]; !ok {
        // Build DSN from individual fields
        dsn = buildDSN(dbConfig)
    }

    db, err := sql.Open(dbType, dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %v", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %v", err)
    }

    return &SQLDatabase{db}, nil
}

func buildDSN(config map[string]string) string {
    switch config["type"] {
    case "mysql":
        return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", 
            config["username"], config["password"], config["host"], config["port"], config["database"])
    case "postgres":
        return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
            config["host"], config["port"], config["username"], config["password"], config["database"])
    default:
        return fmt.Sprintf("%s:%s@%s:%s/%s", 
            config["username"], config["password"], config["host"], config["port"], config["database"])
    }
}



// Manager manages database connections and caches them.
type Manager struct {
    cache       *cache.GenericCache
    mutex       sync.Mutex
}

// NewManager creates a new Manager instance.
func NewManager(cache *cache.GenericCache, vaultConfig *vault.VaultConfig) *Manager {
    return &Manager{
        cache:       cache,
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
            credentials, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "")
            if err != nil {
                return nil, fmt.Errorf("failed to get database credentials: %v", err)
            }

            dbConfig := make(map[string]interface{})

            if dsn, ok := credentials["dsn"].(string); ok {
                dbConfig["type"] = credentials["type"]
                dbConfig["dsn"] = dsn
            } else {
                // Fallback to username/password if DSN is not provided
                dbConfig["type"] = credentials["type"]
                dbConfig["host"] = credentials["host"]
                dbConfig["port"] = credentials["port"]
                dbConfig["username"] = credentials["username"]
                dbConfig["password"] = credentials["password"]
                dbConfig["database"] = credentials["database"]
            }

            dbConn, err = NewDatabaseFactory(dbConfig)
            if err != nil {
                return nil, fmt.Errorf("failed to create database connection: %v", err)
            }

            m.cache.Set(tenantID, dbConn, 1)
        }
    }

    return dbConn.(Database), nil
}


