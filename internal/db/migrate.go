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

func runMigrations(cfg *config.Config, writeDB *sql.DB, isRollback bool) error {
	// MySQL migrations
	driverMySQL, _ := mysql.WithInstance(writeDB, &mysql.Config{})

	mMy, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations/mysql",
		cfg.MySQLConfig.Database, driverMySQL,
	)
	if err != nil {
		return err
	}

	if isRollback {
		if err := mMy.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("mysql migrate: %w", err)
		}
	} else {
		if err := mMy.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("mysql migrate: %w", err)
		}
	}

	return nil
}

func RunMigrations(lc fx.Lifecycle, p RunMigrationsParams) {
	lc.Append(fx.StartHook(func() error {
		return runMigrations(p.Cfg, p.WriteDB, false)
	}))
}

func RunMigrationsVanilla(cfg *config.Config, writeDB *sql.DB, isRollback bool) error {
	return runMigrations(cfg, writeDB, isRollback)
}
