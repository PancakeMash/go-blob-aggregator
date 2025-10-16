package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/PancakeMash/go-blob-aggregator/internal/config"
	"github.com/PancakeMash/go-blob-aggregator/internal/database"
	"github.com/google/uuid"
)

// Functions & Methods
func handlerLogin(s *state, cmd command) error {
	if (len(cmd.args)) == 0 {
		return fmt.Errorf("username is required")
	}

	u, err := s.db.GetUser(context.Background(), cmd.args[0])
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("user not found")
	}
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}
	_ = u

	s.cfg.CurrentUserName = cmd.args[0]
	username := s.cfg.CurrentUserName

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Println("user set to", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("username is required. Usage: gator register name")
	}
	ctx := context.Background()
	username := cmd.args[0]
	id := uuid.New()
	now := time.Now()

	_, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      username,
	})
	if err != nil {
		return err
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("User %q created\n", username)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.ResetUsers(context.Background()); err != nil {
		return err
	}

	fmt.Println("Reset user succeeded")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("no input required. Usage: gator users")
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, u := range users {
		if s.cfg.CurrentUserName == u.Name {
			fmt.Printf("* %s (current)\n", u.Name)
		} else {
			fmt.Printf("* %s\n", u.Name)
		}
	}

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
	db  *database.Queries
	cfg *config.Config
}

type commands struct {
	m map[string]func(*state, command) error
}

type command struct {
	name string
	args []string
}
