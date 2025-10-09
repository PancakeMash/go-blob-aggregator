package main

import (
	"fmt"
	"log"
)

import "github.com/PancakeMash/go-blob-aggregator/internal/config"

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.SetUser("mash")
	if err != nil {
		log.Fatal(err)
	}

	updatedCfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(updatedCfg)
}
