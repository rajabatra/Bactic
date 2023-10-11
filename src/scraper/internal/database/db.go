package database

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"scraper/internal"

	_ "embed"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type BacticDB struct {
	DBConn *sql.DB
	logger log.Logger
}

//go:embed sql/schema.sql
var initSQL string

//go:embed sql/teardown.sql
var teardownSQL string

// Public-facing connection interface to any sql database provided a driver
func NewBacticDB(driverName string, connStr string) BacticDB {
	conn, err := sql.Open(driverName, connStr)
	if err != nil {
		panic(err)
	}

	return BacticDB{
		DBConn: conn,
	}
}

func (b *BacticDB) SetupSchema() {
	_, err := b.DBConn.Exec(initSQL)
	if err != nil {
		panic(err)
	}
}

func (b *BacticDB) TeardownSchema() {
	_, err := b.DBConn.Exec(teardownSQL)
	if err != nil {
		panic(err)
	}
}

func defaultDBConnect() (*BacticDB, error) {

	dbConnURL, found := os.LookupEnv("BACTIC_DB_URL")
	if found == false {
		return &BacticDB{}, errors.New("Unable to find bactic database connection string BACTIC_DB_URL")
	}
	db := NewBacticDB("postgres", dbConnURL)

	return &db, nil
}

// Search the struct for additional data that needs to be crawled before inserting some results
func (db *BacticDB) GetCrawls([]internal.Result) {
	// we are given an athlete id and wish to know if that athlete exists
}

func (db *BacticDB) getAthlete(athID uint32) (internal.Athlete, bool) {
	row := db.DBConn.QueryRow("SELECT id, name, school_id FROM athlete WHERE id = ?", athID)
	var athlete internal.Athlete
	err := row.Scan(&athlete.ID, &athlete.Name, &athlete.SchoolID)
	if err == sql.ErrNoRows {
		return athlete, false
	} else if err != nil {
		log.Fatal("Unable to unmarshal Athlete selection from sql database", err)
	}
	return athlete, true
}

func (db *BacticDB) getSchool(schoolName string) (internal.School, bool) {
	row := db.DBConn.QueryRow("SELECT * FROM school WHERE name = ?", schoolName)
	var school internal.School
	if row.Err() == sql.ErrNoRows {
		return school, false
	} else if row.Err() != nil {
		panic(row.Err())
	} else {
		row.Scan(&school)
		return school, true
	}
}

func (db *BacticDB) insertAthlete(ath internal.Athlete) error {
	// We asume that the athlete's id has already been populated by the tfrrs id
	_, err := db.DBConn.Exec("INSERT INTO athlete(id, name, school_id) values(?, ?, ?)", ath.ID, ath.Name, ath.SchoolID)
	return err
}

func (db *BacticDB) insertSchool(school internal.School) (uint32, error) {
	id := uuid.New().ID()
	school.ID = id
	_, err := db.DBConn.Exec("INSERT INTO school(id, name, division, conference) values(?, ?, ?, ?)", school.ID, school.Name, school.Division, school.Conference)
	if err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (db *BacticDB) insertResult(result internal.Result) error {
	id := uuid.New().ID()
	result.ID = id
	_, err := db.DBConn.Exec(`INSERT INTO result(id, heat_id, ath_id, event_type, pl, date, 
        quant, wind_ms, stage) values(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.ID,
		result.HeatID,
		result.AthleteID,
		result.Place,
		result.Date,
		result.Quantity,
		result.WindMS,
		result.Stage)
	return err
}

// We should process inserts heat-by-heat, since that is how the data is scraped
func (db *BacticDB) InsertHeat(eventType int, meetID uint32, results []internal.Result) {
	// check to see if athlete exists
	heatID := uuid.New().ID()
	_, err := db.DBConn.Exec("INSERT INTO heat(id, meet_id, event_type) value(?, ?, ?)", heatID, meetID, eventType)
	if err != nil {
		db.logger.Println("Error thrown from attempt to create heat:", err)
	}

	// check to see that all schools are in the database

	for _, result := range results {
		err = db.insertResult(result)
		if err != nil {
			db.logger.Println("Error when attempting to insert result into heat:", err)
		}
	}
}

func (db *BacticDB) GetMissingAthletes(results []internal.Result) []uint32 {
	missingID := make([]uint32, 0, len(results))
	for _, result := range results {
		_, found := db.getAthlete(result.AthleteID)
		if found == false {
			missingID = append(missingID, result.AthleteID)
		}
	}
	return missingID
}

func (db *BacticDB) GetMissingSchools(schoolNames []string) []string {
	missingSchool := make([]string, 0, len(schoolNames))
	for _, name := range schoolNames {
		row := db.DBConn.QueryRow("SELECT name FROM school WHERE name = ?", name)
		var queryName string
		err := row.Scan(&queryName)
		if err == sql.ErrNoRows {
			missingSchool = append(missingSchool, name)
		} else if err != nil {
			db.logger.Print("Checking school name failed:", err)
		}
	}
	return missingSchool
}
