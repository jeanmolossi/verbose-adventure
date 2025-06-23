package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeanmolossi/verbose-adventure/internal/core/config"
)

// newPostgreSQL returns a pg connection pool
func newPostgreSQL(dsn string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(context.Background(), `CREATE EXTENSION IF NOT EXISTS "pgcrypto";`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// NewPostgreSQLMaster returns a pg connection pool to master host
func NewPostgreSQLMaster(cfg *config.Config) *pgxpool.Pool {
	db, err := newPostgreSQL(cfg.Database.PostgreSQL.MasterConnString())
	if err != nil {
		panic(err)
	}

	return db
}

// NewPostgreSQLReplica returns a pg connection pool to replica host
func NewPostgreSQLReplica(cfg *config.Config) *pgxpool.Pool {
	db, err := newPostgreSQL(cfg.Database.PostgreSQL.ReplicaConnString())
	if err != nil {
		panic(err)
	}

	return db
}
