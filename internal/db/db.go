// Package db is a package to manage database instructions
package db

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/jeanmolossi/verbose-adventure/internal/config"
)

// newMySQL retorna uma conexão configurada para MySQL.
func newMySQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
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

// NewMySQL retorna uma conexão configurada para MySQL.
func NewMySQL(cfg *config.Config) *sql.DB {
	db, err := newMySQL(cfg.MySQLConfig.WriteDSN())
	if err != nil {
		panic(err)
	}

	return db
}

// NewMySQLRead retorna uma conexão configurada para MySQL na replica de leitura.
func NewMySQLRead(cfg *config.Config) *sql.DB {
	db, err := newMySQL(cfg.MySQLConfig.ReadDSN())
	if err != nil {
		panic(err)
	}

	return db
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
