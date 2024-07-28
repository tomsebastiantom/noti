// package db

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"sync"

// 	"getnoti.com/pkg/cache"
// 	"getnoti.com/pkg/vault"
// 	_ "github.com/denisenkom/go-mssqldb" // Microsoft SQL Server driver
// 	_ "github.com/go-sql-driver/mysql"   // MySQL driver
// 	_ "github.com/lib/pq"                // PostgreSQL driver
// 	_ "modernc.org/sqlite"      // SQLite driver
// )

// // Database interface defines the methods for database operations.
// type Database interface {
// 	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
// 	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
// 	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
// 	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
// 	Prepare(ctx context.Context, query string) (*sql.Stmt, error)
// 	Close() error
// 	Ping(ctx context.Context) error
// }

// // Transaction interface defines methods for database transactions.
// type Transaction interface {
// 	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
// 	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
// 	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
// 	Commit() error
// 	Rollback() error
// }

// // SQLDatabase wraps the sql.DB to implement the Database interface.
// type SQLDatabase struct {
// 	*sql.DB
// }

// // Implement the Database interface methods for SQLDatabase
// func (db *SQLDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
// 	return db.DB.QueryContext(ctx, query, args...)
// }

// func (db *SQLDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
// 	return db.DB.ExecContext(ctx, query, args...)
// }

// func (db *SQLDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
// 	return db.DB.QueryRowContext(ctx, query, args...)
// }

// func (db *SQLDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
// 	tx, err := db.DB.BeginTx(ctx, opts)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &SQLTransaction{tx}, nil
// }

// func (db *SQLDatabase) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
// 	return db.DB.PrepareContext(ctx, query)
// }

// func (db *SQLDatabase) Ping(ctx context.Context) error {
// 	return db.DB.PingContext(ctx)
// }

// // SQLTransaction wraps sql.Tx to implement the Transaction interface
// type SQLTransaction struct {
// 	*sql.Tx
// }

// // Implement the Transaction interface methods for SQLTransaction
// func (tx *SQLTransaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
// 	return tx.Tx.ExecContext(ctx, query, args...)
// }

// func (tx *SQLTransaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
// 	return tx.Tx.QueryContext(ctx, query, args...)
// }

// func (tx *SQLTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
// 	return tx.Tx.QueryRowContext(ctx, query, args...)
// }

// // NewDatabaseFactory initializes a new database connection based on the provided configuration.
// func NewDatabaseFactory(credentials map[string]interface{}) (Database, error) {
// 	var dbConfig map[string]string

// 	if dsn, ok := credentials["dsn"].(string); ok {
// 		dbConfig = map[string]string{
// 			"type": credentials["type"].(string),
// 			"dsn":  dsn,
// 		}
// 	} else {
// 		dbConfig = map[string]string{
// 			"type":     credentials["type"].(string),
// 			"host":     credentials["host"].(string),
// 			"port":     credentials["port"].(string),
// 			"username": credentials["username"].(string),
// 			"password": credentials["password"].(string),
// 			"database": credentials["database"].(string),
// 		}
// 	}

// 	return createDatabaseConnection(dbConfig)
// }

// func createDatabaseConnection(dbConfig map[string]string) (Database, error) {
// 	dbType, ok := dbConfig["type"]
// 	if !ok {
// 		return nil, fmt.Errorf("database type is required")
// 	}

// 	var dsn string
// 	if dsn, ok = dbConfig["dsn"]; !ok {
// 		// Build DSN from individual fields
// 		dsn = buildDSN(dbConfig)
// 	}

// 	db, err := sql.Open(dbType, dsn)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open database: %v", err)
// 	}

// 	if err := db.Ping(); err != nil {
// 		return nil, fmt.Errorf("failed to ping database: %v", err)
// 	}

// 	return &SQLDatabase{db}, nil
// }

// func buildDSN(config map[string]string) string {
// 	switch config["type"] {
// 	case "mysql":
// 		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
// 			config["username"], config["password"], config["host"], config["port"], config["database"])
// 	case "postgres":
// 		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 			config["host"], config["port"], config["username"], config["password"], config["database"])
// 	default:
// 		return fmt.Sprintf("%s:%s@%s:%s/%s",
// 			config["username"], config["password"], config["host"], config["port"], config["database"])
// 	}
// }

// // Manager manages database connections and caches them.
// type Manager struct {
// 	cache *cache.GenericCache
// 	mutex sync.Mutex
// }

// // NewManager creates a new Manager instance.
// func NewManager(cache *cache.GenericCache, vaultConfig *vault.VaultConfig) *Manager {
// 	return &Manager{
// 		cache: cache,
// 	}
// }

// // GetDatabaseConnection retrieves a database connection for the given tenant ID.
// func (m *Manager) GetDatabaseConnection(tenantID string) (Database, error) {
// 	dbConn, found := m.cache.Get(tenantID)
// 	if !found {
// 		m.mutex.Lock()
// 		defer m.mutex.Unlock()

// 		dbConn, found = m.cache.Get(tenantID)
// 		if !found {
// 			credentials, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "")
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to get database credentials: %v", err)
// 			}

// 			dbConfig := make(map[string]interface{})

// 			if dsn, ok := credentials["dsn"].(string); ok {
// 				dbConfig["type"] = credentials["type"]
// 				dbConfig["dsn"] = dsn
// 			} else {
// 				// Fallback to username/password if DSN is not provided
// 				dbConfig["type"] = credentials["type"]
// 				dbConfig["host"] = credentials["host"]
// 				dbConfig["port"] = credentials["port"]
// 				dbConfig["username"] = credentials["username"]
// 				dbConfig["password"] = credentials["password"]
// 				dbConfig["database"] = credentials["database"]
// 			}

// 			dbConn, err = NewDatabaseFactory(dbConfig)
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to create database connection: %v", err)
// 			}

// 			m.cache.Set(tenantID, dbConn, 1)
// 		}
// 	}

// 	return dbConn.(Database), nil
// }
package db

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "sync"

    "getnoti.com/pkg/cache"
    "getnoti.com/pkg/logger"
    "getnoti.com/pkg/vault"
    _ "github.com/denisenkom/go-mssqldb" // Microsoft SQL Server driver
    _ "github.com/go-sql-driver/mysql"   // MySQL driver
    _ "github.com/lib/pq"                // PostgreSQL driver
    sqlite "modernc.org/sqlite"          // SQLite driver
)

func init() {
    sql.Register("sqlite3", &sqlite.Driver{})
    log.Printf("Registered database drivers: %v", sql.Drivers())
}

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
    logger *logger.Logger
}

// Implement the Database interface methods for SQLDatabase
func (db *SQLDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    db.logger.Debug(fmt.Sprintf("Executing query: %s", query))
    return db.DB.QueryContext(ctx, query, args...)
}

func (db *SQLDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    db.logger.Debug(fmt.Sprintf("Executing query: %s", query))
    return db.DB.ExecContext(ctx, query, args...)
}

func (db *SQLDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
    db.logger.Debug(fmt.Sprintf("Executing query: %s", query))
    return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *SQLDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
    db.logger.Debug("Beginning transaction")
    tx, err := db.DB.BeginTx(ctx, opts)
    if err != nil {
        return nil, err
    }
    return &SQLTransaction{tx, db.logger}, nil
}

func (db *SQLDatabase) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
    db.logger.Debug(fmt.Sprintf("Preparing statement: %s", query))
    return db.DB.PrepareContext(ctx, query)
}

func (db *SQLDatabase) Ping(ctx context.Context) error {
    db.logger.Debug("Pinging database")
    return db.DB.PingContext(ctx)
}

// SQLTransaction wraps sql.Tx to implement the Transaction interface
type SQLTransaction struct {
    *sql.Tx
    logger *logger.Logger
}

// Implement the Transaction interface methods for SQLTransaction
func (tx *SQLTransaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    tx.logger.Debug(fmt.Sprintf("Executing query in transaction: %s", query))
    return tx.Tx.ExecContext(ctx, query, args...)
}

func (tx *SQLTransaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    tx.logger.Debug(fmt.Sprintf("Executing query in transaction: %s", query))
    return tx.Tx.QueryContext(ctx, query, args...)
}

func (tx *SQLTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
    tx.logger.Debug(fmt.Sprintf("Executing query in transaction: %s", query))
    return tx.Tx.QueryRowContext(ctx, query, args...)
}

// NewDatabaseFactory initializes a new database connection based on the provided configuration.
func NewDatabaseFactory(credentials map[string]interface{}, logger *logger.Logger) (Database, error) {
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

    return createDatabaseConnection(dbConfig, logger)
}

func createDatabaseConnection(dbConfig map[string]string, logger *logger.Logger) (Database, error) {
    dbType, ok := dbConfig["type"]
    if !ok {
        return nil, fmt.Errorf("database type is required")
    }

    var dsn string
    if dsn, ok = dbConfig["dsn"]; !ok {
        dsn = buildDSN(dbConfig)
    }

    logger.Debug(fmt.Sprintf("Connecting to database of type: %s", dbType))
    db, err := sql.Open(dbType, dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database (type: %s, dsn: %s): %v", dbType, dsn, err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database (type: %s): %v", dbType, err)
    }

    return &SQLDatabase{db, logger}, nil
}

func buildDSN(config map[string]string) string {
    switch config["type"] {
    case "mysql":
        return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
            config["username"], config["password"], config["host"], config["port"], config["database"])
    case "postgres":
        return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
            config["host"], config["port"], config["username"], config["password"], config["database"])
    case "sqlite", "sqlite3":
        return config["database"] // For SQLite, the database is usually just a file path
    default:
        return fmt.Sprintf("%s:%s@%s:%s/%s",
            config["username"], config["password"], config["host"], config["port"], config["database"])
    }
}

// Manager manages database connections and caches them.
type Manager struct {
    cache  *cache.GenericCache
    mutex  sync.Mutex
    logger *logger.Logger
}

// NewManager creates a new Manager instance.
func NewManager(cache *cache.GenericCache, vaultConfig *vault.VaultConfig, logger *logger.Logger) *Manager {
    return &Manager{
        cache:  cache,
        logger: logger,
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
            m.logger.Debug(fmt.Sprintf("Creating new database connection for tenant: %s", tenantID))
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

            dbConn, err = NewDatabaseFactory(dbConfig, m.logger)
            if err != nil {
                return nil, fmt.Errorf("failed to create database connection for tenant %s: %v", tenantID, err)
            }

            m.cache.Set(tenantID, dbConn, 1)
        }
    }

    return dbConn.(Database), nil
}
