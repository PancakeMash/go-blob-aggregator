package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PancakeMash/go-blob-aggregator/internal/config"
	"github.com/PancakeMash/go-blob-aggregator/internal/database"
	"github.com/PancakeMash/go-blob-aggregator/internal/rss"
	"github.com/google/uuid"
)

// Handler Function
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

func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	if len(cmd.args) > 0 {
		url = cmd.args[0]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := rss.FetchFeed(ctx, url)
	if err != nil {
		return err
	}

	b, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(b))

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("require the feed name and url")
	}
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	name := cmd.args[0]
	url := cmd.args[1]

	id := uuid.New()
	now := time.Now()
	userId := user.ID

	result, check := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
		Url:       url,
		UserID:    userId,
	})
	if check != nil {
		return check
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: userId,
		FeedID: result.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("User: %q created \n", user.Name)
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("User ID: %s\n", userId)
	fmt.Printf("Feed: %s\n", result.Name)
	fmt.Printf("Created at %s\n", result.CreatedAt)

	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("no input required. Usage: gator feeds")
	}

	res, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, f := range res {
		fmt.Println(f.FeedName)
		fmt.Println(f.Url)
		fmt.Println(f.UserName)
	}

	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("require only feed url")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no feed with that URL; run addfeed first")
		}
		return err
	}

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	ff, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("%s (%s)\n", ff.FeedName, ff.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	ff, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Println(s.cfg.CurrentUserName)
	for _, u := range ff {
		fmt.Println(u)
	}

	return nil
}

//Methods

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
