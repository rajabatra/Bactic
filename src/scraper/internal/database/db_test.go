package database_test

import (
	"os"
	"scraper/internal"
	"scraper/internal/database"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var db database.BacticDB

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
	athletes := []internal.Result{
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

	missing_ids := db.GetMissingAthletes(athletes)

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

//func TestViolateForeignKey(t *testing.T) {
//    ath := internal.Athlete{
//        ID: 123,
//        Name: "Name",
//        SchoolID: 456,
//    }
//    err := db.InsertAthlete(ath)
//    if err == nil {
//        t.Error("This should violate the foreign key constraint: ", err)
//    }
//}
//
//func TestInsertAthlete(t *testing.T) {
//    school := internal.School{
//        Conference: "Conference",
//        Name: "School",
//        Division: internal.DIII,
//    }
//
//    id, err  := db.InsertSchool(school)
//    if err != nil {
//        t.Error("This should fail on unique constraint: ", err)
//    }
//
//    ath.SchoolID = id
//    err = db.InsertAthlete(ath)
//    if err != nil {
//        t.Error("Insert athlete failed, expected success:", err)
//    }
//}
//
//func TestGetAthlete(t *testing.T) {
//    ath, found := db.GetAthlete(123)
//    if found == false {
//        t.Error("Expected to find athlete 123")
//    }
//
//    t.Log(ath.Name)
//}
