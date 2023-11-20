package database_test

import (
	"testing"
)

// Test that we can create and query a global performance histogram
func TestHistogram(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()

	_, err := db.DBConn.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.DBConn.Exec("INSERT INTO result(id, heat_id, ath_id, pl, quant) VALUES(0, 0, 0, 0, 0)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.DBConn.Exec("INSERT INTO result(id, heat_id, ath_id, pl, quant) VALUES(1, 0, 0, 0, 0.5)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.DBConn.Exec("INSERT INTO result(id, heat_id, ath_id, pl, quant) VALUES(2, 0, 0, 0, 1)")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.DBConn.Exec("INSERT INTO heat(id, meet_id, event_type) VALUES(0, 0, 0)")
	if err != nil {
		t.Fatal(err)
	}

	hist := db.Histogram(0, 3)
	expected := []int{1, 1, 1}
	for i, h := range hist {
		if expected[i] != h {
			t.Fatalf("Expected hist %v but got %v", expected, hist)
		}
	}
}
