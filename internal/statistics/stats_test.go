package statistics_test

import (
	"bactic/internal/data"
	"bactic/internal/database"
	"bactic/internal/statistics"
	"database/sql"
	"math"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDB() *sql.DB {
	db := database.NewBacticDB("postgres", "postgres://postgres:pass@localhost:5432/bactic?sslmode=disable")
	database.TeardownSchema(db)
	database.SetupSchema(db)
	return db
}

func TestStatsInit(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal("could not setup tx")
	}

	ath1 := data.Athlete{
		Name: "Ath1",
		Id:   123,
	}
	ath2 := data.Athlete{
		Name: "Ath2",
		Id:   456,
	}
	err = database.InsertAthlete(tx, ath1)
	if err != nil {
		t.Error("Failed to insert ath1", err)
	}
	err = database.InsertAthlete(tx, ath2)
	if err != nil {
		t.Error("Failed to insert ath2", err)
	}

	heat := []data.Result{
		{
			AthleteId: 123,
			Quantity:  14*60 + 1.29, // Me
			Place:     11,
		},
		{
			AthleteId: 456,
			Quantity:  14*60 + 1.73, // Jack Rosencrans
			Place:     12,
		},
	}

	meet := data.Meet{
		Id:     1234,
		Name:   "Bactic Championships",
		Season: data.OUTDOOR,
		Date:   time.Date(2023, time.May, 6, 0, 0, 0, 0, time.UTC),
	}

	err = database.InsertMeet(tx, meet)
	if err != nil {
		t.Error("Failed to insert preliminary meet", err)
	}

	_, err = database.InsertHeat(tx, data.T5000M, meet.Id, heat)
	if err != nil {
		t.Error("Insert heat operation failed:", err)
	}

	statistics.InitializeAndUpdateDistributions(tx, time.Date(2023, time.May, 7, 0, 0, 0, 0, time.UTC), 1, 0.5)
	if err = tx.Commit(); err != nil {
		t.Error(err)
	}

	row := db.QueryRow("SELECT mean, var, event FROM distns WHERE event = $1;", data.T5000M)
	var (
		mean      float32
		variance  float32
		eventType data.EventType
	)

	if err = row.Scan(&mean, &variance, &eventType); err != nil {
		t.Fatal("scan failed", err)
	}

	expectedMean := (heat[0].Quantity + heat[1].Quantity) / 2
	expectedVariance := math.Pow(float64(heat[0].Quantity-expectedMean), 2) + math.Pow(float64(heat[1].Quantity-expectedMean), 2)
	const epsilon = 1e-4
	if math.Abs(float64(mean-expectedMean)) > epsilon {
		t.Fatal("means did not match")
	}
	if math.Abs(float64(float64(variance)-expectedVariance)) > epsilon {
		t.Fatal("variances did not match")
	}
}

func TestStatsUpdate(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal("could not setup tx")
	}

	ath1 := data.Athlete{
		Name: "Ath1",
		Id:   123,
	}
	ath2 := data.Athlete{
		Name: "Ath2",
		Id:   456,
	}

	err = database.InsertAthlete(tx, ath1)
	if err != nil {
		t.Error("Failed to insert ath1", err)
	}
	err = database.InsertAthlete(tx, ath2)
	if err != nil {
		t.Error("Failed to insert ath2", err)
	}

	heat := []data.Result{
		{
			AthleteId: 123,
			Quantity:  14*60 + 1.29, // Me
			Place:     11,
		},
		{
			AthleteId: 456,
			Quantity:  14*60 + 1.73, // Jack Rosencrans
			Place:     12,
		},
	}

	meet := data.Meet{
		Id:     1234,
		Name:   "Bactic Championships",
		Season: data.OUTDOOR,
		Date:   time.Date(2023, time.May, 6, 0, 0, 0, 0, time.UTC),
	}

	err = database.InsertMeet(tx, meet)
	if err != nil {
		t.Error("Failed to insert preliminary meet", err)
	}

	_, err = database.InsertHeat(tx, data.T5000M, meet.Id, heat)
	if err != nil {
		t.Error("Insert heat operation failed:", err)
	}

	statistics.InitializeAndUpdateDistributions(tx, time.Date(2023, time.May, 7, 0, 0, 0, 0, time.UTC), 1, 0.5)

	// now insert some new result that will invoke the exponential smoothing
	heat1 := []data.Result{
		{
			AthleteId: 123,
			Quantity:  15*60 + 25.6, // Me (lmao)
			Place:     21,
		},
	}

	meet1 := data.Meet{
		Id:     1235,
		Name:   "Division III Outdoor Track & Field Championships",
		Season: data.OUTDOOR,
		Date:   time.Date(2023, time.May, 25, 0, 0, 0, 0, time.UTC),
	}

	err = database.InsertMeet(tx, meet1)
	if err != nil {
		t.Error("Failed to insert preliminary meet", err)
	}

	_, err = database.InsertHeat(tx, data.T5000M, meet1.Id, heat1)
	if err != nil {
		t.Error("Insert heat operation failed:", err)
	}
	// perform the magical update
	const smoothing = 0.5
	statistics.InitializeAndUpdateDistributions(tx, time.Date(2023, time.March, 30, 0, 0, 0, 0, time.UTC), 5, smoothing)
	if err = tx.Commit(); err != nil {
		t.Error(err)
	}

	// validate our results
	row := db.QueryRow("SELECT mean, var, event FROM distns WHERE event = $1;", data.T5000M)
	var (
		mean      float32
		variance  float32
		eventType data.EventType
	)

	if err = row.Scan(&mean, &variance, &eventType); err != nil {
		t.Fatal("scan failed", err)
	}

	expectedMean := (heat[0].Quantity + heat[1].Quantity) / 2
	expectedVariance := math.Pow(float64(heat[0].Quantity-expectedMean), 2) + math.Pow(float64(heat[1].Quantity-expectedMean), 2)
	expectedMean = heat1[0].Quantity*smoothing + (1-smoothing)*expectedMean
	expectedVariance = (1 - smoothing) * expectedVariance

	const epsilon = 1e-4
	if math.Abs(float64(mean-expectedMean)) > epsilon {
		t.Fatalf("means did not match: expected %v, got %v", mean, expectedMean)
	}
	if math.Abs(float64(float64(variance)-expectedVariance)) > epsilon {
		t.Fatalf("variances did not match: expceted %v, got %v", variance, expectedVariance)
	}
}
