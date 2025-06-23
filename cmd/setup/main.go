package main

import (
	"flag"
	"log"

	"github.com/jeanmolossi/verbose-adventure/internal/config"
	coreconfig "github.com/jeanmolossi/verbose-adventure/internal/core/config"
	coredb "github.com/jeanmolossi/verbose-adventure/internal/core/metadata/db"
	"github.com/jeanmolossi/verbose-adventure/internal/db"
)

var rollback bool

func main() {
	flag.BoolVar(&rollback, "rollback", false, "-rollback")

	flag.Parse()

	corecfg, err := coreconfig.New()
	if err != nil {
		panic(err)
	}

	err = coredb.RunMigrationsVanilla(corecfg, rollback)
	if err != nil {
		panic(err)
	}

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	mysql := db.NewMySQL(cfg)

	err = db.RunMigrationsVanilla(cfg, mysql, rollback)
	if err != nil {
		panic(err)
	}

	log.Println("Migrations done!")
}
