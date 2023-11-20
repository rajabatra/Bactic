package database_test

import (
	"bactic/internal"
	"bactic/internal/database"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestGetAthleteID(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	link1 := uuid.New().ID()
	link2 := uuid.New().ID()
	tfrrs := uuid.New().ID()
	bactic := uuid.New().ID()
	fmt.Println(link1, link2, tfrrs, bactic)

	if err := db.AddAthleteRelation(link1, tfrrs); err != nil {
		t.Fatal(err)
	}
	// db.AddAthleteRelation(link2, tfrrs)
	if err := db.AddAthleteRelation(tfrrs, bactic); err != nil {
		t.Fatal(err)
	}

	id, found := db.GetTFRRSAthleteID(link1)
	if !found {
		t.Fatal("Could not find link when there should have been")
	}
	if id != bactic {
		t.Fatal("Found bactic id was not the expected value")
	}
	_, found = db.GetTFRRSAthleteID(link2)
	if found {
		t.Fatal("Link found but should not have been")
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
func TestGetAthlete(t *testing.T) {
	db := setupDummyDB()
	defer db.TeardownSchema()
	ath1 := internal.Athlete{
		Name: "Ath1",
		ID:   123,
	}

	if err := db.InsertAthlete(ath1); err != nil {
		t.Fatal(err)
	}

	res, found := db.GetAthlete(ath1.ID)
	if !found {
		t.Fatal("Could not find athlete in database")
	}

	if res.ID != ath1.ID || res.Name != ath1.Name {
		t.Fatal("Names or IDs did not match")
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
	if err != nil {
		t.Error(err)
	}
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
