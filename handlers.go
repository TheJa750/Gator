package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/TheJa750/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("no username provided")
	}

	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		log.Fatalf("no user: %s exists", username)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Printf("Logged in as: %s\n", username)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("no name provided")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		log.Fatalf("attempted to created user that already exists: %s\n", cmd.args[0])
	}

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}

	u, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("user: %s was created\n", cmd.args[0])

	err = handlerLogin(s, cmd)
	if err != nil {
		return err
	}

	fmt.Println(u)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		log.Fatalf("error resetting database: %v\n", err)
	}

	fmt.Println("Database sucessfully reset")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		log.Fatal("no duration provided.")
	}

	time_between_reqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		log.Fatal("invalid duration string provided")
	}

	fmt.Printf("Collecting feeds every %v\n", time_between_reqs)

	ticker := time.NewTicker(time_between_reqs)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		log.Fatalln("missing name/url arguments")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s sucessfully added.\n", name)

	follow_params := database.CreateFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err = s.db.CreateFollow(context.Background(), follow_params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s is now following %s\n", user.Name, feed.Name)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		user, err := s.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return err
		}

		fmt.Printf("Feed: %s - URL: '%s' - Added by: %s\n", feed.Name, feed.Url, user.Name)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		log.Fatalln("no feed url provided")
	}

	feed, err := s.db.GetFeedFromURL(context.Background(), cmd.args[0])
	if err != nil {
		log.Fatal(err)
	}

	params := database.CreateFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	follow, err := s.db.CreateFollow(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s is now following %s\n", follow.UserName, follow.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		log.Fatalln(err)
	}

	if len(follows) > 0 {
		fmt.Printf("%s is following:\n", user.Name)
	} else {
		fmt.Printf("%s is not following any feeds.\n", user.Name)
	}

	for _, follow := range follows {
		fmt.Println(follow.FeedName)
	}

	return nil
}

func handlerUnFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		log.Fatalln("no feed url provided")
	}

	feed, err := s.db.GetFeedFromURL(context.Background(), cmd.args[0])
	if err != nil {
		log.Fatal(err)
	}

	unfollowParams := database.UnfollowFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.UnfollowFeed(context.Background(), unfollowParams)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s is no longer following %s\n", user.Name, feed.Name)

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) > 0 {
		input, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			log.Fatalln(err)
		}
		limit = input
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		log.Fatalln(err)
	}

	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("==================================================")
	}

	return nil
}
