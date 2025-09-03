package main

import (
	"context"
	"log"
	"sync"
	"time"

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

func scrapeFeed(db *db.Queries, wg *sync.WaitGroup, feed db.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
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
		log.Println("found post", item.Title, "on feed", feed.Name)
	}
	log.Printf("Feed %s collected, %v posts found!", feed.Name, len(rssFeed.Channel.Item))
}
