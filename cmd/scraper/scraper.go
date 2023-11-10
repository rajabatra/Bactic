package main

import (
	"bactic/internal/database"
	"bactic/internal/scrapers/athnet"
	"bactic/internal/scrapers/tfrrs"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	var (
		scrapersList string
		dbURL        string
		found        bool
		verbosity    int
	)
	validScrapers := map[string](func(*database.BacticDB, chan os.Signal, *sync.WaitGroup)){"tfrrs": tfrrs.NewTFRRSCollector, "athnet": athnet.NewAthnetCollector}

	flag.StringVar(&scrapersList, "scrapers", "tfrrs", "Comma-separated list of scrapers to run concurrently. Any of \"tfrrs\" and \"athnet\"")
	flag.StringVar(&dbURL, "db", "", "Fully-qualified postgres url. Overrides the environment variable defined in DB_URL")
	flag.IntVar(&verbosity, "verbosity", 1, "verbosity level (1, 2, 3)")
	flag.Parse()

	if len(dbURL) == 0 {
		dbURL, found = os.LookupEnv("DB_URL")
		if !found {
			log.Fatal("Database url not found in environment variable DB_URL. It must be specified in the arg \"db\"")
		}
	}

	db := database.NewBacticDB("postgres", dbURL)
	db.SetupSchema()
	defer db.Close()

	var scraperSet []func(*database.BacticDB, chan os.Signal, *sync.WaitGroup)

	for _, s := range strings.Split(scrapersList, ",") {
		scraper, found := validScrapers[s]
		if !found {
			log.Fatalf("Passed illegal scraper name %s", s)
		}
		scraperSet = append(scraperSet, scraper)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT)
	var wg sync.WaitGroup

	for _, startScraper := range scraperSet {
		wg.Add(1)
		go startScraper(db, interrupt, &wg)
	}

	go func() {
		<-interrupt
		log.Println("Received interrupt signal, shutting down existing scrapers...")
	}()

	wg.Wait()
	log.Println("All scrapers stopped, closing database...")
	db.Close()
}
