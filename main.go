package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/TheJa750/gator/internal/config"
	"github.com/TheJa750/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	/*
		cfg, err := config.Read()
		for err != nil {
			config.RestoreDefaultCfg()
			cfg, err = config.Read()
		}

		fmt.Println(cfg)
	*/

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	s := state{
		cfg: &cfg,
	}

	db, err := sql.Open("postgres", s.cfg.DbURL)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	dbQueries := database.New(db)

	s.db = dbQueries

	cmds := commands{
		list: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnFollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	input := os.Args

	if len(input) < 2 {
		fmt.Println("not enough arguments provided")
		os.Exit(1)
	}

	cmd := command{
		name: input[1],
		args: input[2:],
	}

	err = cmds.run(&s, cmd)
	if err != nil {
		log.Fatalf("error running command: %v", err)
	}
}
