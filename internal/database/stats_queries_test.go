package database_test

import (
	"bactic/internal"
	"bactic/internal/database"
	"testing"
)

// Test that we can create and query a global performance histogram
func TestHistogram(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)

	_, err := db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO result(id, heat_id, ath_id, pl, quant) VALUES(0, 0, 0, 0, 0)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO result(id, heat_id, ath_id, pl, quant) VALUES(1, 0, 0, 0, 0.5)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO result(id, heat_id, ath_id, pl, quant) VALUES(2, 0, 0, 0, 1)")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO heat(id, meet_id, event_type) VALUES(0, 0, 0)")
	if err != nil {
		t.Fatal(err)
	}

	hist := database.Histogram(db, internal.T5000M, 3)
	expected := []int{1, 1, 1}
	for i, h := range hist {
		if expected[i] != h {
			t.Fatalf("Expected hist %v but got %v", expected, hist)
		}
	}
}
