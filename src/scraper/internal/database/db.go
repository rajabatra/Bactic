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
	logger *log.Logger
}

//go:embed sql/schema.sql
var initSQL string

//go:embed sql/teardown.sql
var teardownSQL string

// Public-facing connection interface to any sql database provided a driver
func NewBacticDB(driverName string, connStr string) *BacticDB {
	conn, err := sql.Open(driverName, connStr)
	if err != nil {
		panic(err)
	}
	logger := log.Default()
	logger.SetPrefix("Database")

	return &BacticDB{
		DBConn: conn,
		logger: logger,
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

	return db, nil
}

// Search the struct for additional data that needs to be crawled before inserting some results
func (db *BacticDB) GetCrawls(results []internal.Result) []uint32 {
	// we are given an athlete id and wish to know if that athlete exists
	toCrawl := make([]uint32, 0, len(results))
	for _, r := range results {
		athlete, found := db.getAthlete(r.ID)
		if !found {
			toCrawl = append(toCrawl, athlete.ID)
		}
	}
	return toCrawl
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

func (db *BacticDB) GetSchool(schoolID uint32) (internal.School, bool) {
	row := db.DBConn.QueryRow("SELECT id, name, division, conference FROM school WHERE id = ?", schoolID)
	var school internal.School
	if row.Err() == sql.ErrNoRows {
		return school, false
	} else if row.Err() != nil {
		panic(row.Err())
	} else {
		row.Scan(&school.ID, &school.Name, &school.Division, &school.Conference)
		return school, true
	}
}

func (db *BacticDB) GetSchoolURL(schoolURL string) (internal.School, bool) {
	row := db.DBConn.QueryRow("SELECT id, name, division, conference FROM school WHERE url = ?", schoolURL)
	var school internal.School

	err := row.Scan(&school.ID, &school.Name, &school.Division, &school.Conference)
	if err == sql.ErrNoRows {
		return school, false
	} else if err != nil {
		panic(err) // determine if this is valid behavior. I believe it is
	}
	return school, true
}

func (db *BacticDB) InsertAthlete(ath internal.Athlete) error {
	// We asume that the athlete's id has already been populated by the tfrrs id
	_, err := db.DBConn.Exec("INSERT INTO athlete(id, name, school_id) values(?, ?, ?)", ath.ID, ath.Name, ath.SchoolID)
	return err
}

func (db *BacticDB) InsertSchool(school internal.School) (uint32, error) {
	id := uuid.New().ID()
	school.ID = id
	_, err := db.DBConn.Exec("INSERT INTO school(id, name, division, conference, url) values(?, ?, ?, ?, ?)", school.ID, school.Name, school.Division, school.Conference, school.URL)
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
func (db *BacticDB) InsertHeat(eventType uint8, meetID uint32, results []internal.Result) {
	// check to see if athlete exists
	heatID := uuid.New().ID()
	_, err := db.DBConn.Exec("INSERT INTO heat(id, meet_id, event_type) VALUES(?, ?, ?)", heatID, meetID, eventType)
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

// For a list of results, check which atheletes are missing and return that list
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

// For a list of school URLs, return a list of those for which there are no matches
func (db *BacticDB) GetMissingSchools(schoolURLs []string) []string {
	missingSchools := make([]string, 0, len(schoolURLs))
	for _, url := range schoolURLs {
		_, found := db.GetSchoolURL(url)
		if !found {
			missingSchools = append(missingSchools, url)
		}
	}
	return missingSchools
}
