package tfrrs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"scraper/internal"
	"scraper/internal/database"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Parse a date string in the tfrrs website format
func parse_date(date string) (time.Time, error) {
	replace_daterange := regexp.MustCompile(`-\d{1,2}`)
	date = string(replace_daterange.ReplaceAll([]byte(date), []byte("")))
	return time.Parse("January 2, 2006", date)
}

func parse_division(region string) int {
	match, err := regexp.MatchString("DIII", region)
	if err != nil {
		panic("Regexp error")
	}
	if match {
		return internal.DIII
	}
	if err != nil {
		panic("Regexp error")
	}
	match, err = regexp.MatchString("DII", region)
	if match {
		return internal.DII
	}
	if err != nil {
		panic("Regexp error")
	}
	match, err = regexp.MatchString("DI", region)
	if match {
		return internal.DI
	}
	if err != nil {
		panic("Regexp error")
	}
	match, err = regexp.MatchString("NAIA", region)
	if match {
		return internal.NAIA
	}
	if err != nil {
		panic("Regexp error")
	}
	return -1
}

// For parsing integers, handles the case where we have a trailing dot from the regexp capture group
func parseInt64(value string) int64 {

	// strip the leading dot or colon
	if len(value) > 0 && (value[len(value)-1] == '.' || value[len(value)-1] == ':') {
		value = value[:len(value)-1]
	}

	valInt, err := strconv.Atoi(value)
	if err != nil {
		return 0
	} else {
		return int64(valInt)
	}
}

func parseTime(t string) (float32, error) {
	if slices.Contains([]string{"DNF", "DQ", "FS", "DNS", "NT"}, t) {
		return 0.0, &internal.TimingError{Name: t}
	} else {
		time_regexp := regexp.MustCompile(`(\d+:)?(\d+).(\d+)`)
		matches := time_regexp.FindStringSubmatch(t)
		if len(matches) != 4 {
			return 0.0, errors.New("Time could not be parsed into expected format string: t")
		}
		minutes := parseInt64(matches[1])
		seconds := parseInt64(matches[2])
		millis := parseInt64(matches[3])

		time_minute := int64(time.Minute)
		time_second := int64(time.Second)
		time_millis := int64(time.Millisecond)
		return float32(time.Duration(minutes*time_minute + seconds*time_second + millis*time_millis).Seconds()), nil
	}
}

var titleToEventEnum = map[string]uint8{
	"5000 meters":       internal.T5000M,
	"5,000 meters":      internal.T5000M,
	"100 meters":        internal.T100M,
	"200 meters":        internal.T200M,
	"400 meters":        internal.T400M,
	"800 meters":        internal.T800M,
	"1500 meters":       internal.T1500M,
	"10,000 meters":     internal.T10000M,
	"110 hurdles":       internal.T110H,
	"400 hurdles":       internal.T400H,
	"3000 steeplechase": internal.T3000S,
	"3000 meters":       internal.T3000M,
	"4 x 100m relay":    internal.T4X100,
	"4 x 100 relay":     internal.T4X100,
	"4 x 400 relay":     internal.T4X400,
	"high jump":         internal.HIGH_JUMP,
	"pole vault":        internal.VAULT,
	"long jump":         internal.LONG_JUMP,
	"triple jump":       internal.TRIPLE_JUMP,
	"shot put":          internal.SHOT,
	"discus":            internal.DISCUS,
	"hammer":            internal.HAMMER,
	"javelin":           internal.JAV,
	"decathlon":         internal.DEC,
	"heptathlon":        internal.HEPT,
	"100 hurdles":       internal.T100H,
}

// Given an event title, return the enumerated event
func parseEvent(event_title string) (uint8, error) {
	sexRe := regexp.MustCompile(`Men's|Women's`)
	trailingRe := regexp.MustCompile(`(Preliminaries|Finals)?(Heat {\d+})?`)
	eventParsed := sexRe.ReplaceAllString(event_title, "")
	eventParsed = trailingRe.ReplaceAllString(eventParsed, "")
	eventParsed = strings.TrimSpace(eventParsed)
	eventParsed = strings.ToLower(eventParsed)

	event_type, exists := titleToEventEnum[eventParsed]
	if exists != true {
		return 0, fmt.Errorf("The event title %s, which converts to key %s could not be mapped to an event", event_title, eventParsed)
	}

	return event_type, nil
}

type ParseResult struct {
	result internal.Result
	err    error
}

func parseResultTable(resultTable [][][]string, logger *log.Logger, eventType uint8) ([]internal.Result, uint8) {
	ret := make([]internal.Result, 0, len(resultTable))

	for _, row := range resultTable {
		result, err := parseResultClass[eventType](row)
		if err != nil {
			logger.Printf("Unable to parse event row due to error: %v. Ignoring", err)
		} else {
			ret = append(ret, result)
		}
	}
	return ret, eventType
}

func parseDistanceResult(row [][]string) (internal.Result, error) {
	if len(row) < 5 {
		return internal.Result{}, fmt.Errorf("Distance result row %v is less than the correct length of 4", row)
	}
	var (
		err   error
		place int
	)

	if len(row[0][0]) > 0 {
		place, err = strconv.Atoi(row[0][0])
		if err != nil {
			return internal.Result{}, err
		}
	}

	time, err := parseTime(row[4][0])
	if err != nil {
		return internal.Result{}, err
	}

	athleteID, err := parseAthleteIDFromURL(row[1][1])
	if err != nil {
		return internal.Result{}, err
	}

	return internal.Result{
		Quantity:  time,
		Place:     uint8(place),
		AthleteID: athleteID,
	}, nil
}

func parseSprintsResult(row [][]string) (internal.Result, error) {
	if len(row) < 5 {
		return internal.Result{}, fmt.Errorf("Distance result row %v is less than the correct length of 4", row)
	}

	var (
		place int
		err   error
	)
	if len(row[0][0]) > 0 {
		place, err = strconv.Atoi(row[0][0])
		if err != nil {
			return internal.Result{}, err
		}
	}

	time, err := parseTime(row[4][0])
	if err != nil {
		return internal.Result{}, err
	}

	athleteID, err := parseAthleteIDFromURL(row[1][1])
	if err != nil {
		return internal.Result{}, err
	}

	return internal.Result{
		Quantity:  time,
		Place:     uint8(place),
		AthleteID: athleteID,
	}, nil
}

func parseAthleteIDFromURL(athleteURL string) (uint32, error) {
	findID := regexp.MustCompile(`https://www.tfrrs.org/athletes/(\d+)`).FindStringSubmatch(athleteURL)
	if len(findID) < 2 {
		return 0, fmt.Errorf("athlete url could not be searched for an id: %s", athleteURL)
	}
	athleteID, err := strconv.Atoi(findID[1])
	if err != nil {
		return 0, fmt.Errorf("Unable to convert athlete url into valid ID: %s", athleteURL)
	}
	return uint32(athleteID), nil
}

func parseRelayResult(row []string) (internal.Result, error) {
	if len(row) < 5 {
		return internal.Result{}, fmt.Errorf("Relay result row %v is less than the correct length of 5", row)
	}

	time, err := parseTime(row[3])
	if err != nil {
		return internal.Result{}, err
	}

	err = nil
	members := make([]uint32, 4)
	for i, m := range strings.Split(row[2], ", ") {
		members[i], err = parseAthleteIDFromURL(m)
	}

	return internal.Result{
		Quantity: time,
		Team:     row[0],
		Members:  members,
	}, nil
}

func notImplementedResult(row [][]string) (internal.Result, error) {
	return internal.Result{}, errors.New("This result parser not yet implemented")
}

var parseResultClass = map[uint8](func([][]string) (internal.Result, error)){
	internal.T5000M:      parseDistanceResult,
	internal.T100M:       parseSprintsResult,
	internal.T200M:       parseSprintsResult,
	internal.T400M:       parseSprintsResult,
	internal.T800M:       parseDistanceResult,
	internal.T1500M:      parseDistanceResult,
	internal.T10000M:     parseDistanceResult,
	internal.T110H:       parseSprintsResult,
	internal.T400H:       parseSprintsResult,
	internal.T3000S:      parseDistanceResult,
	internal.T3000M:      parseDistanceResult,
	internal.T4X100:      notImplementedResult,
	internal.T4X400:      notImplementedResult,
	internal.HIGH_JUMP:   notImplementedResult,
	internal.VAULT:       notImplementedResult,
	internal.LONG_JUMP:   notImplementedResult,
	internal.TRIPLE_JUMP: notImplementedResult,
	internal.SHOT:        notImplementedResult,
	internal.DISCUS:      notImplementedResult,
	internal.HAMMER:      notImplementedResult,
	internal.JAV:         notImplementedResult,
	internal.DEC:         notImplementedResult,
	internal.HEPT:        notImplementedResult,
	internal.T100H:       notImplementedResult,
}

func NewTFRRSCollector(db *database.BacticDB, sig chan os.Signal, wg *sync.WaitGroup) {

	// every day, we check the root page if we have finished scraping for the previous day (hopefully)

	// if channel is signalled, wait for the current scraping meet to finish

	// decrement wg when we are done
	defer wg.Done()
	rootCollector := colly.NewCollector()

	// setup single-page scraper
	meetID := uuid.New().ID()
	meetCollector := NewMeetCollector(db, meetID)

	// Setup the rss feed scraper
	setupRootCollector(rootCollector, meetCollector, db)
}

func setupRootCollector(rootCollector *colly.Collector, meetCollector *colly.Collector, db *database.BacticDB) {
	logger := log.Default()
	logger.SetPrefix("XML Root Collector")

	rootCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Root collector looking at meet RSS", r.URL)
	})

	rootCollector.OnXML("//item", func(x *colly.XMLElement) {
		//TODO: parsing for root node
	})
}

func setupSchoolCollector(schoolCollector *colly.Collector, db *database.BacticDB) {
	logger := log.Default()
	logger.SetPrefix("School Collector")

	schoolCollector.OnRequest(func(r *colly.Request) {
		logger.Println("School collector visiting school", r.URL)
	})

	schoolCollector.OnHTML("h3#team-name", func(h *colly.HTMLElement) {
		titleCaser := cases.Title(language.AmericanEnglish)
		teamName := titleCaser.String(strings.TrimSpace(h.Text))
		division := -1
		var leagues []string
		h.DOM.Parent().Siblings().First().Find("span.panel-heading-normal-text").First().Children().Each(func(i int, s *goquery.Selection) {
			d := parse_division(s.Text())
			if division >= 0 && d >= 0 && division != d {
				logger.Fatalf("Found conflicting divisions in the parsed division list: %d, %d", division, d)
			} else if d >= 0 {
				division = d
			} else {
				leagues = append(leagues, strings.TrimSpace(s.Text()))
			}
		})

		if division < 0 {
			logger.Println("Could not parse a division from the school page", teamName)
		}

		school := internal.School{
			Name:     teamName,
			URL:      h.Request.URL.String(),
			Division: division,
			Leagues:  leagues,
		}
		_, err := db.InsertSchool(school)
		if err != nil {
			logger.Fatal("Error inserting school", err)
		}
	})
}

func setupAthleteCollector(athleteCollector *colly.Collector, db *database.BacticDB) {
	logger := log.Default()
	logger.SetPrefix("Athlete Collector")

	athleteCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Athlete collector visiting athlete", r.URL)
	})

	athleteCollector.OnHTML("h3.panel-title.large-title", func(e *colly.HTMLElement) {
		// what info do we need?
		c := cases.Title(language.AmericanEnglish)
		athName := c.String(strings.Split(strings.TrimSpace(e.Text), "\n")[0])

		// Currently, school is not a foreign key meaning it does not need to be populated on this scrape

		// schoolNameNode := e.DOM.Parent().Siblings().Next()
		// schoolNameURL, exists := schoolNameNode.Attr("href")
		// if exists == false {
		// 	logger.Fatal("Could not find the href attribute in the athlete title line")
		// }
		//schoolNameURL = strings.TrimSpace(schoolNameURL)
		//school, found := db.GetSchoolURL(schoolNameURL)
		//if !found {
		//    schoolCollector.Visit(schoolNameURL)

		//	logger.Fatal("School not found, we should be able to find it:", schoolNameURL)
		//}

		athleteURL := e.Request.URL.String()
		athleteID, err := parseAthleteIDFromURL(athleteURL)
		if err != nil {
			logger.Println(err)
			return
		}

		ath := internal.Athlete{
			ID:   athleteID,
			Name: athName,
		}
		db.InsertAthlete(ath)
	})
}

func NewMeetCollector(db *database.BacticDB, meetID uint32) *colly.Collector {

	logger := log.Default()
	logger.SetPrefix("Meet Collector")

	meetCollector := colly.NewCollector()
	schoolCollector := colly.NewCollector()
	athleteCollector := colly.NewCollector()
	setupAthleteCollector(athleteCollector, db)
	setupSchoolCollector(schoolCollector, db)

	meetCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Visiting meet", r.URL)
	})

	meetCollector.OnHTML("div.row", func(e *colly.HTMLElement) {
		resultsRows := e.DOM.Find("tbody>tr")
		tableLength := resultsRows.Length()
		if tableLength == 0 {
			return
		}

		// the header should be present under other circumstances
		eventType, err := parseEvent(e.DOM.Find("div.custom-table-title>h3").Text())
		if err != nil {
			logger.Println("Unable to parse this table type. Assuming a redundant heat table:", err)
			return
		}

		rowLength := resultsRows.First().Children().Length()
		table := make([][][]string, tableLength)

		athleteURLs := make(map[uint32]string)
		schoolURLs := make([]string, 0, tableLength)
		athleteIDs := make([]uint32, 0, tableLength)
		// Collect into table struct
		resultsRows.Each(func(i int, s *goquery.Selection) {
			table[i] = make([][]string, rowLength)
			s.Children().Each(func(j int, r *goquery.Selection) {
				// strip text and link if it exists
				table[i][j] = make([]string, 0, 2)
				table[i][j] = append(table[i][j], strings.TrimSpace(r.Text()))
				href, found := r.Children().Attr("href")
				if found {
					table[i][j] = append(table[i][j], href)
				}
			})
		})

		// parse all information from table
		resultTable, eventType := parseResultTable(table, logger, eventType)
		logger.Printf("%v, %v", resultTable, eventType)
		athletesToScrape := db.GetMissingAthletes(resultTable)
		schoolsToScrape := db.GetMissingSchools(schoolURLs)

		// visit schools before athletes due to the database relation dependencies
		for _, url := range schoolsToScrape {
			schoolCollector.Visit(url)
		}

		for _, athleteID := range athletesToScrape {
			athleteCollector.Visit(athleteURLs[athleteID])
		}

		// TODO: populate the athlete-school relations
		for i, url := range schoolURLs {
			school, found := db.GetSchoolURL(url)
			if !found {
				logger.Fatal("We must be able to find the school", url)
			}
			db.AddAthleteToSchool(athleteIDs[i], school.ID)
		}

		// once this is done, we can insert the heat
		_, err = db.InsertHeat(eventType, meetID, resultTable)
	})
	return meetCollector
}
