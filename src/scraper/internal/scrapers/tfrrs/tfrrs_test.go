package tfrrs_test

import (
	"scraper/internal/database"
	"scraper/internal/scrapers/tfrrs"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestScraperPage(t *testing.T) {
	db := database.NewBacticDB("sqlite3", ":memory:")
    db.SetupSchema()

    collector := tfrrs.NewMeetCollector(db)
    collector.Visit("https://tfrrs.org/results/79700/m/2023_SCIAC_TF_Championships")

    // assert that we have inserted some values
}



