package db

import (
	"context"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"

	"github.com/jeanmolossi/verbose-adventure/internal/core/config"
)

type RunMigrationsParams struct {
	fx.In

	Cfg *config.Config
}

func runMigrations(cfg *config.Config) error {
	pgPool := NewPostgreSQLMaster(cfg)

	conn, _ := pgPool.Acquire(context.Background())
	defer conn.Release()

	sqldb := stdlib.OpenDB(*pgPool.Config().ConnConfig)

	driverPg, err := postgres.WithInstance(sqldb, &postgres.Config{})
	if err != nil {
		return err
	}

	migrations, err := migrate.NewWithDatabaseInstance(
		"file://internal/core/metadata/db/migrations/postgres",
		cfg.Database.PostgreSQL.Database, driverPg,
	)
	if err != nil {
		return err
	}

	if err := migrations.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func RunMigrations(lc fx.Lifecycle, p RunMigrationsParams) {
	lc.Append(fx.StartHook(func(c context.Context) error {
		return runMigrations(p.Cfg)
	}))
}

func RunMigrationsVanilla(cfg *config.Config) error {
	return runMigrations(cfg)
}
