// These tests should not be run when building for production
package tfrrs_test

import (
	"bactic/internal"
	"bactic/internal/database"
	"bactic/internal/scrapers/tfrrs"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/gocolly/colly"
	_ "github.com/lib/pq"
)

func newTx() *sql.Tx {
	db := database.NewBacticDB("postgres", "postgres://postgres:pass@localhost:5432/bactic?sslmode=disable")
	database.TeardownSchema(db)
	database.SetupSchema(db)

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	return tx
}

func newDB() *sql.DB {
	db := database.NewBacticDB("postgres", "postgres://postgres:pass@localhost:5432/bactic?sslmode=disable")
	database.TeardownSchema(db)
	database.SetupSchema(db)

	return db
}

func TestScraperTFMeet(t *testing.T) {
	tx := newTx()

	meetID := uint32(79700)
	collector := tfrrs.NewMeetCollector()
	database.InsertMeet(tx, internal.Meet{
		ID:     meetID,
		Name:   "2023 SCIAC TF Championships",
		Season: internal.OUTDOOR,
		Date:   time.Date(2023, time.April, 29, 0, 0, 0, 0, time.UTC),
	})
	ctx := colly.NewContext()
	ctx.Put("MeetID", meetID)
	ctx.Put("tx", tx)
	if err := collector.Request("GET", "https://tfrrs.org/results/79700/m/2023_SCIAC_TF_Championships", nil, ctx, nil); err != nil {
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	// TODO: assert that we have inserted some values
}

func TestScraperXCMeet(t *testing.T) {
	tx := newTx()

	meetID := uint32(23293)
	collector := tfrrs.NewMeetCollector()
	database.InsertMeet(tx, internal.Meet{
		ID:     meetID,
		Name:   "2023 SCIAC Cross Country Championships",
		Season: internal.XC,
		Date:   time.Date(2023, time.October, 28, 0, 0, 0, 0, time.UTC),
	})
	ctx := colly.NewContext()
	ctx.Put("MeetID", meetID)
	ctx.Put("tx", tx)
	if err := collector.Request("GET", "https://tfrrs.org/results/xc/23218/2023_SCIAC_Cross_Country_Championships", nil, ctx, nil); err != nil {
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	// assert that we have inserted some values
}

func TestScraperRoot(t *testing.T) {
	db := newDB()
	rss := tfrrs.NewRSSCollector(db)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../../../test/tfrrs_test.rss")
	})
	go http.ListenAndServe(":8080", nil)

	if err := rss.Visit("http://127.0.0.1:8080"); err != nil {
		panic(err)
	}
}
