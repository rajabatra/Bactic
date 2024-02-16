package main

import (
	"bactic/internal/database"
	"bactic/internal/scrapers/tfrrs"
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	var (
		scrapersList string
		scrapeInt    time.Duration
		dbURI        string
		found        bool
		verbosity    int
	)
	validScrapers := map[string](func(*sql.DB, context.Context, *sync.WaitGroup, time.Duration)){
		"tfrrs": tfrrs.NewTFRRSScraper,
		// "athnet": athnet.NewAthnetCollector,
	}

	flag.StringVar(&scrapersList, "scrapers", "tfrrs", "Comma-separated list of scrapers to run concurrently. Any of \"tfrrs\" and \"athnet\"")
	flag.StringVar(&dbURI, "db", "", "Fully-qualified postgres url. Overrides the environment variable defined in DB_URL")
	flag.IntVar(&verbosity, "verbosity", 1, "verbosity level (1, 2, 3)")
	flag.DurationVar(&scrapeInt, "duration", time.Hour*24, "Interval between scrapes")
	flag.Parse()

	log.SetPrefix("Scraper main")

	if len(dbURI) == 0 {
		dbURI, found = os.LookupEnv("DB_URI")
		if !found {
			log.Fatal("Database url not found in environment variable DB_URI. It must be specified in the arg \"db\"")
		}
	}

	db := database.NewBacticDB("postgres", dbURI)
	database.SetupSchema(db)
	defer db.Close()

	var scraperSet []func(*sql.DB, context.Context, *sync.WaitGroup, time.Duration)

	for _, s := range strings.Split(scrapersList, ",") {
		scraper, found := validScrapers[s]
		if !found {
			log.Fatalf("Passed illegal scraper name %s", s)
		}
		scraperSet = append(scraperSet, scraper)
	}

	interrupt := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(interrupt, syscall.SIGINT)
	signal.Notify(interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup

	for _, startScraper := range scraperSet {
		wg.Add(1)
		go startScraper(db, ctx, &wg, scrapeInt)
	}

	go func() {
		<-interrupt
		log.Println("Received interrupt signal, shutting down existing scrapers...")
		cancel()
	}()

	wg.Wait()
	log.Println("All scrapers stopped, closing database...")
	db.Close()
}
