package tfrrs

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

func NewTFRRSScraper(db *sql.DB, ctx context.Context, wg *sync.WaitGroup, scrapeLoop time.Duration) {
	defer wg.Done()

	// every day, we check the root page if we have finished scraping for the previous day (hopefully)
	// if channel is signalled, wait for the current scraping meet to finish
	// decrement wg when we are done

	rssCollector := NewRSSCollector(db, ctx)

	// await for current scraping to finish if interrupt signalled
	scrapeTimer := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-scrapeTimer.C:
			scrapeTimer.Reset(scrapeLoop)
			if err := rssCollector.Visit("https://www.tfrrs.org/results.rss"); err != nil {
				panic(err)
			}
		}
	}
}
