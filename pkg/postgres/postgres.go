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

// NewOrGetSingleton -.
func NewOrGetSingleton(config *config.Config) *Postgres {
	hdlOnce.Do(func() {
		postgres, err := initPg(config)
		if err != nil {
			panic(err)
		}

		pg = postgres
	})
    
	// fmt.Printf("Pg driver: %+v\n", pg)
	return pg
}

func initPg(config *config.Config) (*Postgres, error) {
	pg = &Postgres{
		maxPoolSize:  config.PG.PoolMax,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	poolConfig, err := pgxpool.ParseConfig(config.PG.URL)
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
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
