package database_test

import (
	"scraper/internal"
	"scraper/internal/database"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupDummyDB() *database.BacticDB {
	db := database.NewBacticDB("sqlite3", ":memory:")
	_, err := db.DBConn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		panic(err)
	}
	db.SetupSchema()
	return db
}

func TestTables(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()

	_, err := db.DBConn.Exec("INSERT INTO school(id, name, division, url) VALUES(?, ?, ?, ?)", 2, "school", 3, "abcdef")
	if err != nil {
		t.Error(err)
	}
	_, err = db.DBConn.Exec("INSERT INTO league(school_id, league_name) VALUES(?, ?)", 2, "test")
	if err != nil {
		t.Error(err)
	}
}

func TestGetMissingAthletes(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	heat := []internal.Result{
		{
			ID:        123,
			AthleteID: 123,
		},
		{
			AthleteID: 456,
		},
		{
			AthleteID: 789,
		},
	}

	missing_ids := db.GetMissingAthletes(heat)

	expected := []uint32{123, 456, 789}
	if len(missing_ids) != len(expected) {
		t.Fail()
	}
	for i, exp := range expected {
		if missing_ids[i] != exp {
			t.Fail()
		}
	}
}

func TestInsertSchools(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	school := internal.School{
		Leagues:  []string{"League"},
		Name:     "School",
		Division: internal.DIII,
		URL:      "https://www.tfrrs.org/school_a",
	}

	_, err := db.InsertSchool(school)
	if err != nil {
		t.Error("Unexpected failure to insert:", err)
	}
}

func TestGetMissingSchools(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	schools := []string{"School1", "School2", "School3"}

	missing := db.GetMissingSchools(schools)
	if len(missing) != len(schools) {
		t.Error("Expected lengths are not equal")
	}
	for i, exp := range schools {
		if missing[i] != exp {
			t.Error("The following elements are not equal", missing[i], exp)
		}
	}
}

func TestInsertAthlete(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	ath := internal.Athlete{
		ID:   5,
		Name: "Freddy Fasgi",
	}

	err := db.InsertAthlete(ath)
	if err != nil {
		t.Error("Insert athlete failed, expected success:", err)
	}
}

func TestGetSchoolURL(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	url := "https://www.tfrrs.org/school_b"
	school := internal.School{
		Leagues:  []string{"League1", "League2"},
		Name:     "School",
		Division: internal.DIII,
		URL:      url,
	}

	school_id, err := db.InsertSchool(school)
	if err != nil {
		t.Error("Unexpected failure to school insert: ", err)
	}

	school_ret, found := db.GetSchoolURL(url)
	if !found {
		t.Errorf("Expected to find school with url %s but did not", url)
	}

	if school_ret.ID != school_id {
		t.Errorf("Returned school did not have the same id as inserted")
	}
}

func TestGetSchool(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	school := internal.School{
		Leagues:  []string{"Conference", "League2"},
		Name:     "School",
		Division: internal.DIII,
		URL:      "https://www.tfrrs.org/school_c",
	}

	school_id, err := db.InsertSchool(school)
	if err != nil {
		t.Error("Unexpected failure to school insert: ", err)
	}

	school_ret, found := db.GetSchool(school_id)
	if !found {
		t.Errorf("Expected to find school with id %d but did not", school_id)
	}

	school.URL = ""
	school.ID = school_id
	if school_ret.ID != school.ID {
		t.Errorf("Returned school did not match fields with the inserted value")
	}
}

func TestInsertHeat(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	ath1 := internal.Athlete{
		Name: "Ath1",
		ID:   123,
	}
	ath2 := internal.Athlete{
		Name: "Ath2",
		ID:   456,
	}
	err := db.InsertAthlete(ath1)
	if err != nil {
		t.Error("Failed to insert ath1", err)
	}
	err = db.InsertAthlete(ath2)
	if err != nil {
		t.Error("Failed to insert ath2", err)
	}

	heat := []internal.Result{
		{
			AthleteID: 123,
			Quantity:  14*60 + 1.29, // Me
			Place:     11,
		},
		{AthleteID: 456,
			Quantity: 14*60 + 1.73, // Jack Rosencrans
			Place:    12,
		},
	}

	meet := internal.Meet{
		ID:   1234,
		Name: "Bactic Championships",
		Date: time.Date(2023, time.May, 6, 0, 0, 0, 0, time.UTC),
	}

	err = db.InsertMeet(meet)
	if err != nil {
		t.Error("Failed to insert preliminary meet", err)
	}

	_, err = db.InsertHeat(internal.T5000M, meet.ID, heat)
	if err != nil {
		t.Error("Insert heat operation failed:", err)
	}
}
func TestAthleteSchoolRelation(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	ath1 := internal.Athlete{
		Name: "Ath1",
		ID:   123,
	}
	ath2 := internal.Athlete{
		Name: "Ath2",
		ID:   456,
	}
	err := db.InsertAthlete(ath1)
	if err != nil {
		t.Error("Failed to insert ath1", err)
	}
	err = db.InsertAthlete(ath2)
	if err != nil {
		t.Error("Failed to insert ath2", err)
	}

	school := internal.School{
		Leagues:  []string{"League"},
		Name:     "School",
		Division: internal.DIII,
		URL:      "https://www.tfrrs.org/school_a",
	}

	school.ID, err = db.InsertSchool(school)
	if err != nil {
		t.Error("Unexpected failure to insert:", err)
	}

	err = db.AddAthleteToSchool(ath1.ID, school.ID)
}

func TestInsertMeet(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	meet := internal.Meet{
		ID:   1234,
		Name: "Bactic Championships",
		Date: time.Date(2023, time.May, 6, 0, 0, 0, 0, time.UTC),
	}

	err := db.InsertMeet(meet)
	if err != nil {
		t.Error("Insert meet operation failed", err)
	}
}
