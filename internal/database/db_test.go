package database_test

import (
	"bactic/internal"
	"bactic/internal/database"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func setupTestDB() *sql.DB {
	db := database.NewBacticDB("postgres", "postgres://postgres:pass@localhost:5432/bactic?sslmode=disable")
	database.SetupSchema(db)
	return db
}

func TestTables(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)

	_, err := db.Exec("INSERT INTO school(id, name, division, url) VALUES($1, $2, $3, $4)", 2, "school", 3, "abcdef")
	if err != nil {
		t.Error(err)
	}
	_, err = db.Exec("INSERT INTO league(school_id, league_name) VALUES($1, $2)", 2, "test")
	if err != nil {
		t.Error(err)
	}
}

func TestGetAthleteID(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	link1 := uuid.New().ID()
	link2 := uuid.New().ID()
	tfrrs := uuid.New().ID()
	bactic := uuid.New().ID()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	if err := database.AddAthleteRelation(tx, link1, tfrrs); err != nil {
		t.Fatal(err)
	}
	// db.AddAthleteRelation(link2, tfrrs)
	if err := database.AddAthleteRelation(tx, tfrrs, bactic); err != nil {
		t.Fatal(err)
	}

	id, found := database.GetTFRRSAthleteID(tx, link1)
	if !found {
		t.Fatal("Could not find link when there should have been")
	}
	if id != bactic {
		t.Fatal("Found bactic id was not the expected value")
	}
	_, found = database.GetTFRRSAthleteID(tx, link2)
	if found {
		t.Fatal("Link found but should not have been")
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestInsertSchools(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	school := internal.School{
		Leagues:  []string{"League"},
		Name:     "School",
		Division: internal.DIII,
		URL:      "https://www.tfrrs.org/school_a",
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = database.InsertSchool(tx, school)
	if err != nil {
		t.Error("Unexpected failure to insert:", err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestGetMissingSchools(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	schools := []string{"School1", "School2", "School3"}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	missing := database.GetMissingSchools(tx, schools)
	if len(missing) != len(schools) {
		t.Error("Expected lengths are not equal")
	}
	for i, exp := range schools {
		if missing[i] != exp {
			t.Error("The following elements are not equal", missing[i], exp)
		}
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestInsertAthlete(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	ath := internal.Athlete{
		ID:   5,
		Name: "Freddy Fasgi",
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = database.InsertAthlete(tx, ath)
	if err != nil {
		t.Error("Insert athlete failed, expected success:", err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSchoolURL(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	url := "https://www.tfrrs.org/school_b"
	school := internal.School{
		ID:       uuid.New().ID(),
		Leagues:  []string{"League1", "League2"},
		Name:     "School",
		Division: internal.DIII,
		URL:      url,
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = database.InsertSchool(tx, school)
	if err != nil {
		t.Error("Unexpected failure to school insert: ", err)
	}

	school_ret, found := database.GetSchoolURL(tx, url)
	if !found {
		t.Errorf("Expected to find school with url %s but did not", url)
	}

	if school_ret.ID != school.ID {
		t.Errorf("Returned school did not have the same id as inserted")
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSchool(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	school := internal.School{
		ID:       uuid.New().ID(),
		Leagues:  []string{"Conference", "League2"},
		Name:     "School",
		Division: internal.DIII,
		URL:      "https://www.tfrrs.org/school_c",
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = database.InsertSchool(tx, school)
	if err != nil {
		t.Error("Unexpected failure to school insert: ", err)
	}

	school_ret, found := database.GetSchool(tx, school.ID)
	if !found {
		t.Errorf("Expected to find school with id %d but did not", school.ID)
	}

	if school_ret.ID != school.ID {
		t.Errorf("Returned school did not match fields with the inserted value")
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
func TestGetAthlete(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	ath1 := internal.Athlete{
		Name: "Ath1",
		ID:   123,
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.InsertAthlete(tx, ath1); err != nil {
		t.Fatal(err)
	}

	res, found := database.GetAthlete(tx, ath1.ID)
	if !found {
		t.Fatal("Could not find athlete in database")
	}

	if res.ID != ath1.ID || res.Name != ath1.Name {
		t.Fatal("Names or IDs did not match")
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
func TestInsertHeat(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	ath1 := internal.Athlete{
		Name: "Ath1",
		ID:   123,
	}
	ath2 := internal.Athlete{
		Name: "Ath2",
		ID:   456,
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = database.InsertAthlete(tx, ath1)
	if err != nil {
		t.Error("Failed to insert ath1", err)
	}
	err = database.InsertAthlete(tx, ath2)
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

	err = database.InsertMeet(tx, meet)
	if err != nil {
		t.Error("Failed to insert preliminary meet", err)
	}

	_, err = database.InsertHeat(tx, internal.T5000M, meet.ID, heat)
	if err != nil {
		t.Error("Insert heat operation failed:", err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
func TestAthleteSchoolRelation(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	ath1 := internal.Athlete{
		Name: "Ath1",
		ID:   123,
	}

	ath2 := internal.Athlete{
		Name: "Ath2",
		ID:   456,
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = database.InsertAthlete(tx, ath1)
	if err != nil {
		t.Fatal("Failed to insert ath1", err)
	}
	err = database.InsertAthlete(tx, ath2)
	if err != nil {
		t.Fatal("Failed to insert ath2", err)
	}

	school := internal.School{
		ID:       uuid.New().ID(),
		Leagues:  []string{"League"},
		Name:     "School",
		Division: internal.DIII,
		URL:      "https://www.tfrrs.org/school_a",
	}

	err = database.InsertSchool(tx, school)
	if err != nil {
		t.Fatal("Unexpected failure to insert:", err)
	}

	err = database.AddAthleteToSchool(tx, ath1.ID, school.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestInsertMeet(t *testing.T) {
	db := setupTestDB()
	defer database.TeardownSchema(db)
	meet := internal.Meet{
		ID:   1234,
		Name: "Bactic Championships",
		Date: time.Date(2023, time.May, 6, 0, 0, 0, 0, time.UTC),
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = database.InsertMeet(tx, meet)
	if err != nil {
		t.Error("Insert meet operation failed", err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
