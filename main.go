package main

import (
	"fmt"
	"log"
	"os"
)

import "github.com/PancakeMash/go-blob-aggregator/internal/config"

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	appState := &state{config: &cfg}

	cliCommands := &commands{}
	cliCommands.m = make(map[string]func(*state, command) error)
	cliCommands.register("login", handlerLogin)

	if len(os.Args) < 2 {
		log.Fatal("not enough arguments were provided")
	}

	commandName := os.Args[1]
	commandArgs := os.Args[2:]

	err = cliCommands.run(appState, command{commandName, commandArgs})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(appState.config)
}
