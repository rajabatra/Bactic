package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"scraper/internal"

	_ "embed"

	"github.com/google/uuid"
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
		log.Fatalf("Could not connect to database with url %s due to error: %v", connStr, err)
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
	row := db.DBConn.QueryRow("SELECT id, name FROM athlete WHERE id = ?", athID)
	var athlete internal.Athlete
	err := row.Scan(&athlete.ID, &athlete.Name)
	if err == sql.ErrNoRows {
		return athlete, false
	} else if err != nil {
		log.Fatal("Unable to unmarshal Athlete selection from sql database", err)
	}

	rows, err := db.DBConn.Query("SELECT school_id FROM athlete_in_school WHERE athlete_id = ?", athID)
	defer rows.Close()
	if err != nil && err != sql.ErrNoRows {
		log.Fatal("Query to athlete-school-relation table failed", err)
	}
	var schools []uint32
	var school uint32
	for rows.Next() {
		err = row.Scan(&school)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Fatal("Unable to unmarshal school id", err)
		}
		schools = append(schools, school)
	}
	athlete.Schools = schools

	return athlete, true
}

func (db *BacticDB) GetSchool(schoolID uint32) (internal.School, bool) {
	row := db.DBConn.QueryRow("SELECT id, name, division FROM school WHERE id = ?", schoolID)
	var school internal.School
	if row.Err() == sql.ErrNoRows {
		return school, false
	} else if row.Err() != nil {
		panic(row.Err())
	}

	row.Scan(&school.ID, &school.Name, &school.Division)
	leagues, err := db.DBConn.Query("SELECT league_name FROM league WHERE school_id = ?", school.ID)
	defer leagues.Close()
	if row.Err() == sql.ErrNoRows {
		return school, false
	} else if err != nil {
		panic(err)
	}

	var league string
	var l []string
	for leagues.Next() {
		leagues.Scan(&league)
		l = append(l, league)
	}
	school.Leagues = l
	return school, true
}

func (db *BacticDB) GetSchoolURL(schoolURL string) (internal.School, bool) {
	var school internal.School
	row := db.DBConn.QueryRow("SELECT id, name, division FROM school WHERE url = ?", schoolURL)

	err := row.Scan(&school.ID, &school.Name, &school.Division)
	if err == sql.ErrNoRows {
		return school, false
	} else if err != nil {
		panic(err)
	}

	leagues, err := db.DBConn.Query("SELECT league_name FROM league WHERE school_id = ?", school.ID)
	defer leagues.Close()
	if err == sql.ErrNoRows {
		return school, true
	} else if err != nil {
		panic(err)
	}

	var league string
	var l []string
	for leagues.Next() {
		leagues.Scan(&league)
		l = append(l, league)
	}
	school.Leagues = l
	return school, true
}

func (db *BacticDB) InsertAthlete(ath internal.Athlete) error {
	// We assume that the athlete's id has already been populated by the tfrrs id
	_, err := db.DBConn.Exec("INSERT INTO athlete(id, name) VALUES(?, ?)", ath.ID, ath.Name)
	if err != nil {
		return err
	}
	for _, schoolID := range ath.Schools {
		err = db.AddAthleteToSchool(ath.ID, schoolID)
		if err != nil {
			return fmt.Errorf("Error creating athlete school relation: %s", err)
		}
	}

	return nil
}

func (db *BacticDB) GetAthlete(athID uint32) (internal.Athlete, bool) {
	row := db.DBConn.QueryRow("SELECT name FROM athlete WHERE id = ?", athID)
	var ath internal.Athlete
	if row.Err() == sql.ErrNoRows {
		return ath, false
	}
	err := row.Scan(&ath.Name)
	if err != nil {
		panic(err)
	}
	ath.ID = athID
	return ath, true
}

func (db *BacticDB) AddAthleteToSchool(athID uint32, schoolID uint32) error {
	row := db.DBConn.QueryRow("SELECT school_id from athlete_in_school WHERE school_id = ? and athlete_id = ?", schoolID, athID)
	if err := row.Scan(); err == sql.ErrNoRows {
		_, err := db.DBConn.Exec("INSERT INTO athlete_in_school(athlete_id, school_id) VALUES(?, ?)", athID, schoolID)
		return err
	}
	return nil
	//row := db.DBConn.QueryRow("SELECT * FROM athlete_in_school WHERE athlete_id = ? AND school_id = ?", athID, schoolID)
	//if row.Err() != nil && row.Err() != sql.ErrNoRows {
	//	log.Fatal("Threw unexpected error ", row.Err())
	//	return nil
	//} else if row.Err() == sql.ErrNoRows {
	//	log.Println("Could not find any rows, adding to table")
	//	_, err := db.DBConn.Exec("INSERT INTO athlete_in_school(athlete_id, school_id) VALUES(?, ?)", athID, schoolID)
	//	return err
	//} else {
	//	log.Println("Row already exists")
	//	return nil
	//}
}

func (db *BacticDB) InsertSchool(school internal.School) (uint32, error) {
	id := uuid.New().ID()
	school.ID = id
	cur, err := db.DBConn.Begin()
	if err != nil {
		return 0, err
	}
	_, err = cur.Exec("INSERT INTO school(id, name, division, url) VALUES(?, ?, ?, ?)", school.ID, school.Name, school.Division, school.URL)
	if err != nil {
		cur.Rollback()
		return 0, err
	}

	for _, league := range school.Leagues {
		_, err := cur.Exec("INSERT INTO league(school_id, league_name) VALUES(?, ?)", school.ID, league)
		if err != nil {
			cur.Rollback()
			return 0, err
		}
	}

	if err = cur.Commit(); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func insertResult(cur *sql.Tx, result internal.Result) error {
	id := uuid.New().ID()
	result.ID = id
	_, err := cur.Exec(`INSERT INTO result(id, heat_id, ath_id, pl, 
        quant, wind_ms, stage) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		result.ID,
		result.HeatID,
		result.AthleteID,
		result.Place,
		result.Quantity,
		result.WindMS,
		result.Stage)
	return err
}

// We should process inserts heat-by-heat, since that is how the data is scraped
func (db *BacticDB) InsertHeat(eventType uint8, meetID uint32, results []internal.Result) (uint32, error) {
	// check to see if athlete exists
	heatID := uuid.New().ID()
	cur, err := db.DBConn.Begin()
	if err != nil {
		return 0, err
	}
	_, err = cur.Exec("INSERT INTO heat(id, meet_id, event_type) VALUES(?, ?, ?)", heatID, meetID, eventType)
	if err != nil {
		cur.Rollback()
		return 0, fmt.Errorf("Could not create heat table: %s", err)
	}

	// check to see that all schools are in the database

	for _, result := range results {
		result.HeatID = heatID
		if err = insertResult(cur, result); err != nil {
			cur.Rollback()
			return 0, fmt.Errorf("Could not insert result: %s", err)
		}
	}

	if err = cur.Commit(); err != nil {
		cur.Rollback()
		return 0, fmt.Errorf("Could not commit table insert: %s", err)
	} else {
		return heatID, nil
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

func (db *BacticDB) InsertMeet(meet internal.Meet) error {
	_, err := db.DBConn.Exec("INSERT INTO meet(id, name, date) VALUES(?, ?, ?)", meet.ID, meet.Name, meet.Date)
	return err
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

func (db *BacticDB) Close() {
	db.DBConn.Close()
}
