package main

import (
	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"github.com/jeanmolossi/verbose-adventure/internal/db"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.New, // *config.Config

			fx.Annotate(
				db.NewMySQL, // *sql.DB (MySQL Write)
				fx.ResultTags(`name:"mysqlMaster"`),
			),
		),
		fx.Invoke(
			db.RunMigrations,
		),
	)
}
