package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/PancakeMash/go-blob-aggregator/internal/database"
)
import _ "github.com/lib/pq"
import "github.com/PancakeMash/go-blob-aggregator/internal/config"

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	appState := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cliCommands := &commands{}
	cliCommands.m = make(map[string]func(*state, command) error)
	cliCommands.register("login", handlerLogin)
	cliCommands.register("register", handlerRegister)
	cliCommands.register("reset", handlerReset)

	if len(os.Args) < 2 {
		log.Fatal("not enough arguments were provided")
	}

	commandName := os.Args[1]
	commandArgs := os.Args[2:]

	err = cliCommands.run(appState, command{commandName, commandArgs})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(appState.cfg)
}
