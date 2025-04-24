package db

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jeanmolossi/verbose-adventure/internal/config"
	_ "github.com/lib/pq"
)

// NewMySQL retorna uma conexão configurada para MySQL.
func NewMySQL(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.MySQLConfig.WriteDSN())
	if err != nil {
		return nil, err
	}

	// Pool tuning
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping inicial
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// NewPostgres retorna uma conexão configurada para PostgreSQL.
func NewPostgres(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.PGConfig.WriteDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
