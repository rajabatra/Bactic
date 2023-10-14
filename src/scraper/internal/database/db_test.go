package database_test

import (
	"os"
	"scraper/internal"
	"scraper/internal/database"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var db *database.BacticDB

func TestMain(m *testing.M) {
	db = database.NewBacticDB("sqlite3", ":memory:")
	_, err := db.DBConn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		panic(err)
	}

	db.SetupSchema()

	output := m.Run()

	db.TeardownSchema()
	os.Exit(output)
}

func TestGetMissingAthletes(t *testing.T) {
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

func TestGetMissingSchools(t *testing.T) {
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

func TestViolateForeignKey(t *testing.T) {
	ath := internal.Athlete{
		ID:       123,
		Name:     "Name",
		SchoolID: 456,
	}
	err := db.InsertAthlete(ath)
	if err == nil {
		t.Error("This should violate the foreign key constraint: ", err)
	}
}

func TestInsertAthlete(t *testing.T) {
	school := internal.School{
		Conference: "Conference",
		Name:       "School",
		Division:   internal.DIII,
		URL:        "https://www.tfrrs.org/school_a",
	}

	school_id, err := db.InsertSchool(school)
	if err != nil {
		t.Error("Unexpected failure to school insert: ", err)
	}
	ath := internal.Athlete{
		ID:       5,
		Name:     "Freddy Fasgi",
		SchoolID: school_id,
	}

	err = db.InsertAthlete(ath)
	if err != nil {
		t.Error("Insert athlete failed, expected success:", err)
	}
}

func TestGetSchoolURL(t *testing.T) {
	url := "https://www.tfrrs.org/school_b"
	school := internal.School{
		Conference: "Conference",
		Name:       "School",
		Division:   internal.DIII,
		URL:        url,
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
	school := internal.School{
		Conference: "Conference",
		Name:       "School",
		Division:   internal.DIII,
		URL:        "https://www.tfrrs.org/school_c",
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
	if school_ret != school {
		t.Errorf("Returned school did not match fields with the inserted value")
	}
}

func TestInsertHeat(t *testing.T) {
	heat := []internal.Result{
		{
			AthleteID: 123,
			HeatID:    1,
			Quantity:  14*60 + 1.29, // Me
			Place:     11,
			Date:      time.Date(2023, time.May, 6, 0, 0, 0, 0, time.UTC),
		},
		{AthleteID: 456,
			HeatID:   1,
			Quantity: 14*60 + 1.73, // Jack Rosencrans
			Place:    12,
			Date:     time.Date(2023, time.May, 6, 0, 0, 0, 0, time.UTC),
		},
	}

    id := uuid.New().ID()

    db.InsertHeat(internal.T5000M, id, heat)
}
