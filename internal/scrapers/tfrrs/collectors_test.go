package tfrrs_test

import (
	"bactic/internal"
	"bactic/internal/database"
	"bactic/internal/scrapers/tfrrs"
	"testing"
	"time"

	"github.com/gocolly/colly"
	_ "github.com/mattn/go-sqlite3"
)

func newDemoDB() *database.BacticDB {
	db := database.NewBacticDB("sqlite3", ":memory:")
	_, err := db.DBConn.Exec("PRAGMA foreign_keys=true")
	if err != nil {
		panic(err)
	}
	db.SetupSchema()

	return db

}

func TestScraperTFMeet(t *testing.T) {
	db := newDemoDB()

	meetID := uint32(79700)
	collector := tfrrs.NewTFRRSTrackCollector(db)
	db.InsertMeet(internal.Meet{
		ID:     meetID,
		Name:   "2023 SCIAC TF Championships",
		Season: internal.OUTDOOR,
		Date:   time.Date(2023, time.April, 29, 0, 0, 0, 0, time.UTC),
	})
	ctx := colly.NewContext()
	ctx.Put("MeetID", meetID)
	if err := collector.Request("GET", "https://tfrrs.org/results/79700/m/2023_SCIAC_TF_Championships", nil, ctx, nil); err != nil {
		t.Fatal(err)
	}

	// assert that we have inserted some values
}

func TestScraperXCMeet(t *testing.T) {
	db := newDemoDB()

	meetID := uint32(23218)
	collector := tfrrs.NewTFRRSXCCollector(db)
	db.InsertMeet(internal.Meet{
		ID:     meetID,
		Name:   "2023 SCIAC Cross Country Championships",
		Season: internal.XC,
		Date:   time.Date(2023, time.October, 28, 0, 0, 0, 0, time.UTC),
	})
	ctx := colly.NewContext()
	ctx.Put("MeetID", meetID)
	if err := collector.Request("GET", "https://tfrrs.org/results/xc/23218/2023_SCIAC_Cross_Country_Championships", nil, ctx, nil); err != nil {
		t.Fatal(err)
	}

	// assert that we have inserted some values
}

func TestScraperRoot(t *testing.T) {
	db := newDemoDB()
	rss := tfrrs.NewRSSCollector(db)
	if err := rss.Visit("https://www.tfrrs.org/results.rss"); err != nil {
		panic(err)
	}
}
