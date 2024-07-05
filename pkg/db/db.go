package db

import (
	"fmt"
	"getnoti.com/config"
	"context"
	"getnoti.com/pkg/db/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Database interface
type Database interface {
	Close() error
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row

}
// NewDatabaseFactory creates a new database connection based on the given config
func NewDatabaseFactory(cfg *config.Config) (Database, error) {
	var database Database
	var err error

	switch cfg.Database.Type {
	case "postgres":
		database, err = postgres.New(cfg)
	case "mysql":
		// database, err = mysql.New(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return database, nil
}
