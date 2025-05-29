package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	"getnoti.com/pkg/logger"

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
	logger logger.Logger
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
	logger logger.Logger
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
func NewDatabaseFactory(credentials map[string]interface{}, logger logger.Logger) (Database, error) {
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

func createDatabaseConnection(dbConfig map[string]string, logger logger.Logger) (Database, error) {
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

func createTenantDatabase(tenantID string, dbConfig map[string]interface{}, logger logger.Logger) (Database, map[string]interface{}, error) {
	dsn, ok := dbConfig["dsn"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("invalid DSN in configuration")
	}

	dbType, ok := dbConfig["type"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("invalid database type in configuration")
	}

	err := createDatabase(dbType, dsn, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database for tenant %s: %v", tenantID, err)
	}

	updatedDSN := updateDSNWithTenant(dbType, dsn, tenantID)

	updatedConfig := map[string]interface{}{
		"type": dbType,
		"dsn":  updatedDSN,
	}

	newDBConnection, err := NewDatabaseFactory(updatedConfig, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new database connection: %v", err)
	}

	return newDBConnection, updatedConfig, nil
}

func createDatabase(dbType, dsn, tenantID string) error {
	switch dbType {
	case "sqlite":
		// Create a new SQLite database file for the tenant
		tenantDBFile := fmt.Sprintf("%s.db", tenantID)
		db, err := sql.Open(dbType, tenantDBFile)
		if err != nil {
			return fmt.Errorf("failed to open connection: %v", err)
		}
		defer db.Close()

	case "postgres", "cockroach":
		db, err := sql.Open(dbType, dsn)
		if err != nil {
			return fmt.Errorf("failed to open connection: %v", err)
		}
		defer db.Close()

		query := fmt.Sprintf("CREATE DATABASE %s", tenantID)
		_, err = db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create database: %v", err)
		}

	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	return nil
}

func updateDSNWithTenant(dbType, dsn, tenantID string) string {

	if dbType == "sqlite" {
		return fmt.Sprintf("%s.db", tenantID)
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return dsn // Return original DSN if parsing fails
	}

	switch dbType {
	case "mysql":
		path := strings.TrimLeft(u.Path, "/")
		u.Path = fmt.Sprintf("/%s", tenantID)
		if path != "" {
			u.RawQuery = fmt.Sprintf("%s&%s", path, u.RawQuery)
		}
	case "postgres", "cockroach":
		u.Path = fmt.Sprintf("/%s", tenantID)
	}

	return u.String()
}


