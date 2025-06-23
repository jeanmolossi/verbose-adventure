// Package db is a package to manage metadata databse operations
package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	"github.com/jeanmolossi/verbose-adventure/internal/core/config"
)

// newPostgreSQL returns a pg connection pool
func newPostgreSQL(dsn string, options ...ConfigOption) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		opt(poolConfig)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// NewPostgreSQLMaster returns a pg connection pool to master host
func NewPostgreSQLMaster(cfg *config.Config, options ...ConfigOption) *pgxpool.Pool {
	db, err := newPostgreSQL(cfg.Database.PostgreSQL.MasterConnString(), options...)
	if err != nil {
		panic(err)
	}

	return db
}

// NewPostgreSQLReplica returns a pg connection pool to replica host
func NewPostgreSQLReplica(cfg *config.Config, options ...ConfigOption) *pgxpool.Pool {
	db, err := newPostgreSQL(cfg.Database.PostgreSQL.ReplicaConnString(), options...)
	if err != nil {
		panic(err)
	}

	return db
}
