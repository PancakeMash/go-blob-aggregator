package main

import (
	"fmt"

	"github.com/PancakeMash/go-blob-aggregator/internal/config"
)

// Functions & Methods
func handlerLogin(s *state, cmd command) error {
	if (len(cmd.args)) == 0 {
		return fmt.Errorf("username is required")
	}

	s.config.CurrentUserName = cmd.args[0]
	username := s.config.CurrentUserName

	if err := s.config.SetUser(username); err != nil {
		return err
	}

	fmt.Println("user set to", username)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	cmdHandler, ok := c.m[cmd.name]
	if !ok {
		return fmt.Errorf("command %s not found", cmd.name)
	}

	err := cmdHandler(s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.m[name] = f
}

//Types

type state struct {
	config *config.Config
}

type commands struct {
	m map[string]func(*state, command) error
}

type command struct {
	name string
	args []string
}
