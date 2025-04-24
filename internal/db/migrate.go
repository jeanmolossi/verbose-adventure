package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"go.uber.org/fx"
)

func RunMigrations(lc fx.Lifecycle, mysqlDB, pg *sql.DB, cfg *config.Config) {
	lc.Append(fx.StartHook(func() error {
		// MySQL migrations
		driverMySql, _ := mysql.WithInstance(mysqlDB, &mysql.Config{})
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
	}))
}
