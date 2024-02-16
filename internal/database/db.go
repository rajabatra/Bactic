package database

import (
	"bactic/internal/data"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "embed"

	"github.com/google/uuid"
)

//go:embed sql/schema.sql
var initSQL string

//go:embed sql/teardown.sql
var teardownSQL string

// Public-facing connection interface to any sql database provided a driver
func NewBacticDB(driverName string, connStr string) *sql.DB {
	conn, err := sql.Open(driverName, connStr)
	if err != nil {
		log.Panicf("could not connect to database with url %s due to error: %v", connStr, err)
	}
	SetupSchema(conn)
	return conn
}

func SetupSchema(db *sql.DB) {
	_, err := db.Exec(initSQL)
	if err != nil {
		panic(err)
	}
}

func TeardownSchema(db *sql.DB) {
	_, err := db.Exec(teardownSQL)
	if err != nil {
		panic(err)
	}
}

// Search the struct for additional data that needs to be crawled before inserting some results
func GetCrawls(tx *sql.Tx, results []data.Result) []uint32 {
	// we are given an athlete id and wish to know if that athlete exists
	toCrawl := make([]uint32, 0, len(results))
	for _, r := range results {
		athlete, found := getAthlete(tx, r.Id)
		if !found {
			toCrawl = append(toCrawl, athlete.Id)
		}
	}
	return toCrawl
}

func getAthlete(tx *sql.Tx, athID uint32) (data.Athlete, bool) {
	row := tx.QueryRow("SELECT id, name FROM athlete WHERE id = $1", athID)
	var athlete data.Athlete
	err := row.Scan(&athlete.Id, &athlete.Name)
	if err == sql.ErrNoRows {
		return athlete, false
	} else if err != nil {
		log.Panic("Unable to unmarshal Athlete selection from sql database", err)
	}

	rows, err := tx.Query("SELECT school_id FROM athlete_in_school WHERE athlete_id = $1", athID)
	if err != nil && err != sql.ErrNoRows {
		log.Panic("Query to athlete-school-relation table failed", err)
	}
	var schools []uint32
	var school uint32
	for rows.Next() {
		err = row.Scan(&school)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			log.Panic("Unable to unmarshal school id", err)
		}
		schools = append(schools, school)
	}
	if err = rows.Close(); err != nil {
		panic(err)
	}
	athlete.Schools = schools

	return athlete, true
}

func GetSchool(tx *sql.Tx, schoolID uint32) (data.School, bool) {
	row := tx.QueryRow("SELECT id, name, division FROM school WHERE id = $1", schoolID)
	var school data.School
	if row.Err() == sql.ErrNoRows {
		return school, false
	} else if row.Err() != nil {
		panic(row.Err())
	}

	row.Scan(&school.Id, &school.Name, &school.Division)
	leagues, err := tx.Query("SELECT league_name FROM league WHERE school_id = $1", school.Id)
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

	if err = leagues.Close(); err != nil {
		panic(err)
	}

	school.Leagues = l
	return school, true
}

func GetSchoolURL(tx *sql.Tx, schoolURL string) (data.School, bool) {
	var school data.School
	row := tx.QueryRow("SELECT id, name, division FROM school WHERE url = $1", schoolURL)

	err := row.Scan(&school.Id, &school.Name, &school.Division)
	if err == sql.ErrNoRows {
		return school, false
	} else if err != nil {
		panic(err)
	}

	leagues, err := tx.Query("SELECT league_name FROM league WHERE school_id = $1", school.Id)
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

	if err = leagues.Close(); err != nil {
		panic(err)
	}
	school.Leagues = l
	return school, true
}

func InsertAthlete(tx *sql.Tx, ath data.Athlete) error {
	// We assume that the athlete's id has already been populated by the tfrrs id
	_, err := tx.Exec("INSERT INTO athlete(id, name) VALUES($1, $2)", ath.Id, ath.Name)
	if err != nil {
		return err
	}
	for _, schoolID := range ath.Schools {
		err = AddAthleteToSchool(tx, ath.Id, schoolID)
		if err != nil {
			return fmt.Errorf("could not create athlete school relation: %s", err)
		}
	}
	return nil
}

// Get athlete struct from database according to bactic athlete id
func GetAthlete(tx *sql.Tx, athID uint32) (data.Athlete, bool) {
	row := tx.QueryRow("SELECT name FROM athlete WHERE id = $1", athID)
	var ath data.Athlete
	err := row.Scan(&ath.Name)
	if err == sql.ErrNoRows {
		return ath, false
	} else if err != nil {
		panic(err)
	}

	ath.Id = athID
	// TODO: get other information about athelete (year, schools)
	return ath, true
}

func AddAthleteToSchool(tx *sql.Tx, athID uint32, schoolID uint32) error {
	var s uint32
	err := tx.QueryRow("SELECT school_id from athlete_in_school WHERE school_id = $1 AND athlete_id = $2", schoolID, athID).Scan(&s)
	if err == sql.ErrNoRows {
		_, err := tx.Exec("INSERT INTO athlete_in_school(athlete_id, school_id) VALUES($1, $2)", athID, schoolID)
		return err
	}
	return err
}

func InsertSchool(tx *sql.Tx, school data.School) error {
	_, err := tx.Exec("INSERT INTO school(id, name, division, url) VALUES($1, $2, $3, $4)", school.Id, school.Name, school.Division, school.Url)
	if err != nil {
		return err
	}

	for _, league := range school.Leagues {
		_, err := tx.Exec("INSERT INTO league(school_id, league_name) VALUES($1, $2)", school.Id, league)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertResult(tx *sql.Tx, result data.Result) error {
	result.Id = uuid.New().ID()
	_, err := tx.Exec(`INSERT INTO result(id, heat_id, ath_id, pl,
        quant, wind_ms, stage) VALUES($1, $2, $3, $4, $5, $6, $7)`,
		result.Id,
		result.HeatId,
		result.AthleteId,
		result.Place,
		result.Quantity,
		result.WindMs,
		result.Stage)
	return err
}

// We should process inserts heat-by-heat, since that is how the data is scraped
func InsertHeat(tx *sql.Tx, eventType data.EventType, meetID uint32, results []data.Result) (uint32, error) {
	// check to see if athlete exists
	heatID := uuid.New().ID()
	_, err := tx.Exec("INSERT INTO heat(id, meet_id, event_type) VALUES($1, $2, $3)", heatID, meetID, eventType)
	if err != nil {
		return 0, err
	}

	// check to see that all schools are in the database

	for _, result := range results {
		result.HeatId = heatID
		if err := insertResult(tx, result); err != nil {
			return 0, err
		}
	}

	return heatID, nil
}

// Query the athlete map table for an id relation
func GetAthleteRelation(tx *sql.Tx, id uint32) (uint32, bool) {
	row := tx.QueryRow("SELECT y FROM athlete_map WHERE x = $1", id)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false
	} else if err != nil {
		panic(err)
	} else {
		return id, true
	}
}

// For a link id, return the bactic athlete id if it can be found
func GetTFRRSAthleteID(tx *sql.Tx, linkID uint32) (bacticID uint32, found bool) {
	tfrrs, found := GetAthleteRelation(tx, linkID)
	if !found {
		return 0, false
	}
	bactic, found := GetAthleteRelation(tx, tfrrs)
	if !found {
		return tfrrs, true
	}
	return bactic, true
}

func AddAthleteRelation(tx *sql.Tx, x uint32, y uint32) error {
	_, err := tx.Exec("INSERT INTO athlete_map(x, y) VALUES($1, $2)", x, y)
	return err
}

func InsertMeet(tx *sql.Tx, meet data.Meet) error {
	_, err := tx.Exec("INSERT INTO meet(id, name, date, season) VALUES($1, $2, $3, $4)", meet.Id, meet.Name, meet.Date, meet.Season)
	return err
}

// For a list of school URLs, return a list of those for which there are no matches
func GetMissingSchools(tx *sql.Tx, schoolURLs []string) []string {
	missingSchools := make([]string, 0, len(schoolURLs))
	for _, url := range schoolURLs {
		_, found := GetSchoolURL(tx, url)
		if !found {
			missingSchools = append(missingSchools, url)
		}
	}
	return missingSchools
}

func getEventHistory(tx *sql.Tx, id uint32) map[data.EventType]data.Result {
	rows, err := tx.Query(`SELECT pl, quant, wind_ms
		FROM result
		WHERE result.ath_id = $1`, id)
	if err != nil {
		log.Fatalf("error querying event history: %s", err)
	}

	for rows.Next() {
		rows.
	}

}

func GetAthleteSummary(tx *sql.Tx, id uint32) (summary data.AthleteSummary, found bool) {
	s := data.AthleteSummary{}
	a, found := GetAthlete(tx, id)
	if !found {
		return s, false
	}
	s.Athlete = a
	s.Results = getEventHistory(tx, id)
	return s, true
}
