package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/TheJa750/gator/internal/database"
	"github.com/google/uuid"
)

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	xml_data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	feedData := &RSSFeed{}
	err = xml.Unmarshal(xml_data, feedData)
	if err != nil {
		return nil, err
	}

	feedData.Channel.Title = html.UnescapeString(feedData.Channel.Title)
	feedData.Channel.Description = html.UnescapeString(feedData.Channel.Description)

	for i, item := range feedData.Channel.Item {
		feedData.Channel.Item[i].Title = html.UnescapeString(item.Title)
		feedData.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return feedData, nil
}

func scrapeFeeds(s *state) {
	fmt.Println("")
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("Pulling feed for: %s\n", nextFeed.Name)
	fmt.Printf("Time pulled: %v\n", time.Now())

	params := database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: time.Now(),
		ID:        nextFeed.ID,
	}

	err = s.db.MarkFeedFetched(context.Background(), params)
	if err != nil {
		log.Println(err)
	}

	rssFeed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		log.Println(err)
	}

	for _, item := range rssFeed.Channel.Item {
		pub_at := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			pub_at = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}

		if !pub_at.Valid {
			if t, err := time.Parse(time.RFC1123, item.PubDate); err == nil {
				pub_at = sql.NullTime{
					Time:  t,
					Valid: true,
				}
			}
		}

		desc := sql.NullString{
			String: item.Description,
			Valid:  true,
		}

		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: desc,
			PublishedAt: pub_at,
			FeedID:      nextFeed.ID,
		}

		_, err := s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				continue
			}
			log.Printf("Couldn't create post: %v\n", err)
			continue
		}
	}
	log.Printf("Feed %s collected, %v posts found\n", nextFeed.Name, len(rssFeed.Channel.Item))

}
