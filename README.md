# Feed AggreGATOR

## Requirements

- [Postgres](https://www.postgresql.org/download/)
- [Go](https://go.dev/doc/install)

## Installation

`go install github.com/TheJa750/Gator@latest`

### Where is gator installed?

After running `go install github.com/TheJa750/Gator@latest`, the `gator` binary will be placed in your Go binary directory.

- By default, this is usually `$HOME/go/bin` (for example, `/home/your-username/go/bin/gator` on Linux/Mac, or `C:\Users\your-username\go\bin\gator` on Windows).
- If you have set the environment variable `GOBIN`, the binary will be installed there instead.
- You can check your Go binary directory by running `go env GOBIN` and `go env GOPATH` in your terminal.

To use `gator` from anywhere, make sure your binary directory is in your system `PATH`. See [here](https://golangdocs.com/gopath-and-goroot-in-go-programming) for help.

## Configuration

### Config initialization
When you run `gator` for the first time, it creates a default config file at `~/.gatorconfig.json`.  
Open that file and replace the example database URL with your actual Postgres connection string. For example:

```json
{
    "DbURL": "postgres://username:password@localhost:5432/dbname?sslmode=disable"
}
```

After updating the config, you can use gator commands normally.

### Database migrations

To set up your database tables, you need to run the migrations located [here](https://github.com/TheJa750/Gator/tree/main/sql/schema).

Simple way to accomplish this is to:
1. Download or clone the [`sql/schema`](https://github.com/TheJa750/Gator/tree/main/sql/schema)
2. [Install goose](https://github.com/pressly/goose#install)
3. Open a terminal and navigate to the directory containing the `.sql` migration files.
4. Run the migration command (replace the connection string with your own from the config file):
```sh
goose postgres "postgres://username:password@localhost:5432/dbname?sslmode=disable" up
```
This will apply all of the migrations in that directory to set up the database tables.

## Commands/Usage

### Getting Started
Now that everything is installed and the database tables are in place, you are ready to create a user and begin adding feeds.
1. Register a user. This can be done by running:

```sh
gator register [name]
```

2. You should receive the messages: "user: [name] was created." and "Logged in as: [name]" Now you are able to add feeds.

```sh
gator addfeed "Feed Name" "Feed xml URL"
```

You can use any name for the feed, but the URL needs to match the source you are trying to aggregate.

3. You are now ready to begin aggregating the feed(s). Using:

```sh
gator agg [interval]
```

Where `[interval]` is the time between fetching the most recent feeds from the source and is formatted as `1s`, `1m`, `1h` or a combination (e.g. `1h37m22s`).
*Tip: Beconsiderate and avoid overly frequent fetch intervals to prevent spamming source servers.* You can leave this terminal running in the background to continuously update with new posts from each feed. Use a new terminal window to run other commands while aggregation is ongoing..

4. Browse your posts:

```sh
gator browse [num_feeds]
```

The optional argument `[num_feeds]`(defaults to 2) controls how many posts to display. The `browse` command  will display the title, brief description and the link to the full article.

### More Commands
The full list of commands:

```sh
gator register <name>
#required name, creates new user and logs in as user

gator login <name>
#required name, logins into existing user

gator users
#lists all existing users and shows current user

gator agg <interval>
#required interval, format 1h3m50s, continuously fetches feeds (immediately, then every interval)

gator addfeed <name> <url>
#required name and url, adds a new feed to database, automatically follows for current user

gator feeds
#lists all feeds currently in database and which user added them

gator follow <url>
#required url, used to get updates from a feed added by another user

gator following
#lists all feeds current user is following

gator unfollow <url>
#required url, current user will no longer recieve updates from feed

gator browse [num_feeds]
#optional num_feeds(default: 2), displays given number of posts from followed feeds.

gator reset
#WARNING: This will premanently erase all users, feeds, and posts!
```