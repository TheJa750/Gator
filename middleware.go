package main

import (
	"context"
	"log"

	"github.com/TheJa750/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			log.Fatalf("unable to get user id for: %s\n", s.cfg.CurrentUserName)
		}

		return handler(s, cmd, user)
	}
}
