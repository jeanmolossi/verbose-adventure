package main

import (
	"log"

	"github.com/jeanmolossi/verbose-adventure/internal/config"
	coreconfig "github.com/jeanmolossi/verbose-adventure/internal/core/config"
	coredb "github.com/jeanmolossi/verbose-adventure/internal/core/metadata/db"
	"github.com/jeanmolossi/verbose-adventure/internal/db"
)

func main() {
	corecfg, err := coreconfig.New()
	if err != nil {
		panic(err)
	}

	err = coredb.RunMigrationsVanilla(corecfg)
	if err != nil {
		panic(err)
	}

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	mysql := db.NewMySQL(cfg)

	err = db.RunMigrationsVanilla(cfg, mysql)
	if err != nil {
		panic(err)
	}

	log.Println("Migrations done!")
}
