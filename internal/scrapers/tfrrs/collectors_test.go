package tfrrs_test

import (
	"bactic/internal/database"
	"bactic/internal/scrapers/tfrrs"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func TestScraperMeet(t *testing.T) {
	db := database.NewBacticDB("sqlite3", "sciac.db")
	_, err := db.DBConn.Exec("PRAGMA foreign_keys=true")
	if err != nil {
		t.Fatalf("Failed to set foreign keys pragma in test database: %v", err)
	}
	db.SetupSchema()

	collector := tfrrs.NewTFRRSTrackCollector(db, uuid.New().ID())
	collector.Visit("https://tfrrs.org/results/79700/m/2023_SCIAC_TF_Championships")

	// assert that we have inserted some values
}

func TestScraperRoot(t *testing.T) {
	// db := database.NewBacticDB("sqlite3", ":memory:")
}
