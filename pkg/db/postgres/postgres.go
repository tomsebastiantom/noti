// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"getnoti.com/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Pool *pgxpool.Pool
}

var pg *Postgres
var hdlOnce sync.Once


func New(config *config.Config) (*Postgres, error) {
	var err error
	hdlOnce.Do(func() {
		var postgres *Postgres
		postgres, err = initPg(config)
		if err != nil {
			return
		}
		pg = postgres
	})

	return pg, err
}


func initPg(config *config.Config) (*Postgres, error) {
	pg = &Postgres{
		maxPoolSize:  config.Database.Postgres.PoolMax,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	poolConfig, err := pgxpool.ParseConfig(config.Database.Postgres.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	}

	return pg, nil
}

// Close -.
func (p *Postgres) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
		return nil
	}
	return fmt.Errorf("no connection pool to close")
}

// Query -.
func (p *Postgres) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return p.Pool.Query(ctx, query, args...)
}

// Exec -.
func (p *Postgres) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return p.Pool.Exec(ctx, query, args...)
}

// QueryRow -.
func (p *Postgres) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return p.Pool.QueryRow(ctx, query, args...)
}
