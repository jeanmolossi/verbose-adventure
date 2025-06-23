package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/fx"

	"github.com/jeanmolossi/verbose-adventure/internal/config"
)

type RunMigrationsParams struct {
	fx.In

	WriteDB *sql.DB `name:"mysqlMaster"`
	Cfg     *config.Config
}

func FirstMigration(p RunMigrationsParams) {
	cfg := p.Cfg

	_, err := p.WriteDB.Exec(`SELECT 1 FROM identity_providers LIMIT 1`)
	if err == nil {
		return
	}

	// MySQL migrations
	driverMySql, err := mysql.WithInstance(p.WriteDB, &mysql.Config{})
	if err != nil {
		panic(err)
	}

	mMy, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations/mysql",
		cfg.MySQLConfig.Database, driverMySql,
	)
	if err != nil {
		panic(err)
	}

	if err := mMy.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
}

func runMigrations(cfg *config.Config, writeDB *sql.DB) error {
	// MySQL migrations
	driverMySql, _ := mysql.WithInstance(writeDB, &mysql.Config{})

	mMy, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations/mysql",
		cfg.MySQLConfig.Database, driverMySql,
	)
	if err != nil {
		return err
	}

	if err := mMy.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("mysql migrate: %w", err)
	}

	// PostgreSQL migrations
	// driverPg, _ := postgres.WithInstance(pg, &postgres.Config{})
	// mPg, err := migrate.NewWithDatabaseInstance(
	// 	"file://internal/db/migrations/postgres",
	// 	cfg.PGConfig.Database, driverPg,
	// )
	// if err != nil {
	// 	return err
	// }
	//
	// if err := mPg.Up(); err != nil && err != migrate.ErrNoChange {
	// 	return fmt.Errorf("postgres migrate: %w", err)
	// }

	return nil
}

func RunMigrations(lc fx.Lifecycle, p RunMigrationsParams) {
	lc.Append(fx.StartHook(func() error {
		return runMigrations(p.Cfg, p.WriteDB)
	}))
}

func RunMigrationsVanilla(cfg *config.Config, writeDB *sql.DB) error {
	return runMigrations(cfg, writeDB)
}
