package tfrrs

import (
	"bactic/internal/database"
	"context"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/google/uuid"
)

func NewTFRRSScraper(db *database.BacticDB, ctx context.Context, wg *sync.WaitGroup, scrapeLoop time.Duration) {
	defer wg.Done()

	// every day, we check the root page if we have finished scraping for the previous day (hopefully)

	// if channel is signalled, wait for the current scraping meet to finish

	// decrement wg when we are done
	rootCollector := colly.NewCollector()

	// setup single-page scraper
	meetID := uuid.New().ID()
	meetCollector := NewTFRRSTrackCollector(db, meetID)

	// Setup the rss feed scraper
	setupRSSCollector(rootCollector, meetCollector, db)

	// await for current scraping to finish if interrupt signalled
	scrapeTimer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-scrapeTimer.C:
			scrapeTimer.Reset(scrapeLoop)
			rootCollector.Visit("https://www.tfrrs.org/results.rss")
		}
	}
}
