package tfrrs_test

import (
<<<<<<< HEAD
	"bactic/internal/database"
	"bactic/internal/scrapers/tfrrs"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func TestScraperMeet(t *testing.T) {
	db := database.NewBacticDB("sqlite3", "sciac.db")
=======
	"bactic/internal"
	"bactic/internal/database"
	"bactic/internal/scrapers/tfrrs"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestScraperTFMeet(t *testing.T) {
	db := database.NewBacticDB("sqlite3", ":memory:")
>>>>>>> 115597e (file splitting and stats files)
	_, err := db.DBConn.Exec("PRAGMA foreign_keys=true")
	if err != nil {
		t.Fatalf("Failed to set foreign keys pragma in test database: %v", err)
	}
	db.SetupSchema()

<<<<<<< HEAD
	collector := tfrrs.NewTFRRSTrackCollector(db, uuid.New().ID())
=======
	collector := tfrrs.NewTFRRSTrackCollector(db, 79700)
	db.InsertMeet(internal.Meet{
		ID:     79700,
		Name:   "2023 SCIAC TF Championships",
		Season: internal.OUTDOOR,
		Date:   time.Date(2023, time.April, 29, 0, 0, 0, 0, time.UTC),
	})
>>>>>>> 115597e (file splitting and stats files)
	collector.Visit("https://tfrrs.org/results/79700/m/2023_SCIAC_TF_Championships")

	// assert that we have inserted some values
}

<<<<<<< HEAD
=======
func TestScraperXCMeet(t *testing.T) {
	db := database.NewBacticDB("sqlite3", "sciacxc.db")
	_, err := db.DBConn.Exec("PRAGMA foreign_keys=true")
	if err != nil {
		t.Fatalf("Failed to set foreign keys pragma in test database: %v", err)
	}
	db.SetupSchema()

	collector := tfrrs.NewTFRRSTrackCollector(db, 23218)
	db.InsertMeet(internal.Meet{
		ID:     23218,
		Name:   "2023 SCIAC Cross Country Championships",
		Season: internal.XC,
		Date:   time.Date(2023, time.October, 28, 0, 0, 0, 0, time.UTC),
	})
	collector.Visit("https://tfrrs.org/results/xc/23218/2023_SCIAC_Cross_Country_Championships")

	// assert that we have inserted some values
}

>>>>>>> 115597e (file splitting and stats files)
func TestScraperRoot(t *testing.T) {
	// db := database.NewBacticDB("sqlite3", ":memory:")
}
