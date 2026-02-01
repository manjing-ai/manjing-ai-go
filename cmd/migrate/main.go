package main

import (
	"fmt"
	"os"

	"manjing-ai-go/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: migrate [up|down|force <version>]")
		os.Exit(1)
	}

	cfg := config.MustLoad()
	m, err := migrate.New("file://migrations", cfg.DB.DSN)
	if err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			panic(err)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			panic(err)
		}
	case "force":
		if len(os.Args) < 3 {
			fmt.Println("usage: migrate force <version>")
			os.Exit(1)
		}
		version := os.Args[2]
		var v uint
		if _, err := fmt.Sscanf(version, "%d", &v); err != nil {
			panic(err)
		}
		if err := m.Force(int(v)); err != nil {
			panic(err)
		}
	default:
		fmt.Println("unknown command")
		os.Exit(1)
	}
}
