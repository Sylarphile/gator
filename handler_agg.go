package main

import (
	"context"
	"fmt"
	"time"
	"strings"
	"log"
	"database/sql"
	
	"github.com/Sylarphile/gator/internal/database"
	"github.com/google/uuid"
)

func handlerAgg(s* state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <time_between_requests>", cmd.Name)
	}
	time_between_reqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("couldn't parse time: %w", err)
	}
	fmt.Printf("Collecting feeds every %d\n", time_between_reqs)
	scrapeFeed(s.db)

	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		scrapeFeed(s.db)
	}
}

func scrapeFeed(db *database.Queries) {
	next_feed, err := db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Printf("couldn't get next feed: %v", err)
	}

	err = db.MarkFeedFetched(context.Background(), next_feed.ID)
	if err != nil {
		log.Printf("couldn't mark feed: %v", err)
	}

	feedRSS, err := fetchFeed(context.Background(), next_feed.Url)
	if err != nil {
		log.Printf("couldn't fetch RSS feed: %v", err)
	}

	for _, item := range feedRSS.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			FeedID:    next_feed.ID,
			Title:     item.Title,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Couldn't create post: %v", err)
			continue
		}
	}
	log.Printf("Feed %s collected, %v posts found", next_feed.Name, len(feedRSS.Channel.Item))
}