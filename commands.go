package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/PancakeMash/go-blob-aggregator/internal/config"
	"github.com/PancakeMash/go-blob-aggregator/internal/database"
	"github.com/PancakeMash/go-blob-aggregator/internal/rss"
	"github.com/google/uuid"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			return fmt.Errorf("not logged in")
		}
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func scrapeFeeds(s *state) error {
	lf, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	ff, err := rss.FetchFeed(context.Background(), lf.Url)
	if err != nil {
		return err
	}

	for _, item := range ff.Channel.Item {
		fmt.Println(item.Title)
	}

	return nil

}

// Handler Functions
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
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: agg <time_between_reqs>, e.g. '1m' or '10s'")
	}

	d, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", cmd.args[0], err)
	}

	fmt.Printf("Collecting feeds every %s\n", d)

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	// run immediately, then on each tick
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Printf("scrape error: %v\n", err)
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("require the feed name and url")
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

	_, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
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

func handlerFollow(s *state, cmd command, user database.User) error {
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

func handlerFollowing(s *state, cmd command, user database.User) error {

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

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("require only feed url")
	}
	url := cmd.args[0]
	urlId, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	res, err := s.db.UnfollowFeed(context.Background(), database.UnfollowFeedParams{
		UserID: user.ID,
		FeedID: urlId.ID,
	})
	if err != nil {
		return err
	}

	_ = res

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
