package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pranayyb/whisp-rss-aggregator/internal/db"
)

func startScraping(db *db.Queries, concurrency int, timeBetweenRequest time.Duration) {
	log.Printf("scraping on %v goroutines every %s duration", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurrency),
		)
		if err != nil {
			log.Println("error fetching feeds: ", err)
			continue
		}
		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()
	}
}

func scrapeFeed(database *db.Queries, wg *sync.WaitGroup, feed db.Feed) {
	defer wg.Done()

	_, err := database.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("Error marking feed as fetched: ", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Println("error Fetching feed: ", err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		// log.Println("found post", item.Title, "on feed", feed.Name)
		description := sql.NullString{}
		if item.Description != "" {
			description = sql.NullString{
				String: item.Description,
				Valid:  true,
			}
		}
		pubAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Println("error parsing time: ", err)
			continue
		}
		_, err = database.CreatePost(context.Background(), db.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: description,
			PublishedAt: pubAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				// log.Println("post already exists, skipping")
				continue
			}
			log.Println("error creating post: ", err)
			continue
		}
	}
	log.Printf("Feed %s collected, %v posts found!", feed.Name, len(rssFeed.Channel.Item))
}
