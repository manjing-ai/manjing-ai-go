package main

import (
	"flag"
	"fmt"
	"os"

	"manjing-ai-go/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	configPath := fs.String("config", "", "config file path")
	_ = fs.Parse(os.Args[1:])
	args := fs.Args()
	if len(args) < 1 {
		fmt.Println("usage: migrate [-config path] [up|down|force <version>]")
		os.Exit(1)
	}

	var cfg *config.Config
	if *configPath != "" {
		cfg = config.MustLoadWithPath(*configPath)
	} else {
		cfg = config.MustLoad()
	}
	m, err := migrate.New("file://migrations", cfg.DB.DSN)
	if err != nil {
		panic(err)
	}

	switch args[0] {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			panic(err)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			panic(err)
		}
	case "force":
		if len(args) < 2 {
			fmt.Println("usage: migrate [-config path] force <version>")
			os.Exit(1)
		}
		version := args[1]
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
