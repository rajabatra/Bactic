package tfrrs

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"scraper/internal"
	"scraper/internal/database"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/exp/slices"
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
	"4 x 400 relay":     internal.T4X400,
	"high jump":         internal.HIGH_JUMP,
	"pole vault":        internal.VAULT,
	"long jump":         internal.LONG_JUMP,
	"triple jump":       internal.TRIPLE_JUMP,
	"shot put":          internal.SHOT,
	"discus":            internal.DISCUS,
	"hammer throw":      internal.HAMMER,
	"javelin":           internal.JAV,
	"decathlon":         internal.DEC,
	"heptathlon":        internal.HEPT,
	"100 hurdles":       internal.T100H,
}

// Given an event title, return the enumerated event
func parseEvent(event_title string) (uint8, error) {
	sex_re := regexp.MustCompile(`Men's|Women's`)
	trailing_re := regexp.MustCompile(`(Preliminaries|Finals)?(Heat {\d+})?`)
	event_parsed := sex_re.ReplaceAllString(event_title, "")
	event_parsed = trailing_re.ReplaceAllString(event_parsed, "")
	event_parsed = strings.TrimSpace(event_parsed)
	event_parsed = strings.ToLower(event_parsed)

	event_type, err := titleToEventEnum[event_parsed]
	if err != true {
		return 0, fmt.Errorf("The event title %s, which converts to key %s could not be mapped to an event", event_title, event_parsed)
	}

	return event_type, nil
}

type ParseResult struct {
	result internal.Result
	err    error
}

func parseResultTable(resultTable [][]string, logger *log.Logger) ([]internal.Result, uint8) {
	eventType, err := parseEvent(resultTable[0][0])
	if err != nil {
		logger.Printf("Unable to parse event title: %s", err)
		return []internal.Result{}, 0
	}

	ret := make([]internal.Result, 0, len(resultTable)-1)

	for i := 1; i < len(resultTable); i++ {
		result, err := parseResultRow(resultTable[i], eventType)
		if err != nil {
			logger.Print("Unable to parse event row", err)
		} else {
			ret = append(ret, result)
		}
	}

	return ret, eventType
}

func parseDistanceResult(row []string) (internal.Result, error) {
	if len(row) < 5 {
		return internal.Result{}, fmt.Errorf("Distance result row %v is less than the correct length of 4", row)
	}

	time, err := parseTime(row[4])
	if err != nil {
		return internal.Result{}, err
	}

	return internal.Result{
		Quantity: time,
	}, nil
}

func parseSprintsResult(row []string) (internal.Result, error) {
	if len(row) < 5 {
		return internal.Result{}, fmt.Errorf("Distance result row %v is less than the correct length of 4", row)
	}

	time, err := parseTime(row[4])
	if err != nil {
		return internal.Result{}, err
	}

	return internal.Result{
		Quantity: time,
	}, nil
}

func notImplementedResult(row []string) (internal.Result, error) {
	return internal.Result{}, errors.New("This result parser not yet implemented")
}

var parseResultClass = map[uint8](func([]string) (internal.Result, error)){
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

func parseResultRow(resultRow []string, resultType uint8) (internal.Result, error) {
	place, err := strconv.Atoi(resultRow[0])
	if err != nil {
		return internal.Result{},
			fmt.Errorf("Unable to parse place string %e", err)
	}

	athleteID, err := strconv.Atoi(resultRow[1])
	if err != nil {
		return internal.Result{},
			fmt.Errorf("Unable to parse parse athlete id string %e", err)
	}

	row, err := parseResultClass[resultType](resultRow)
	if err != nil {
		return internal.Result{}, err
	}

	row.AthleteID = uint32(athleteID)
	row.Place = uint8(place)
	return row, nil
}

type TFRRSParser *colly.Collector

func NewTFRRSCollector(db *database.BacticDB) TFRRSParser {
	rootCollector := colly.NewCollector()

	// setup single-page scraper
	meetCollector := NewMeetCollector(db)

	// Setup the rss feed scraper
	setupRootCollector(rootCollector, meetCollector, db)
	return rootCollector
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
		teamName := h.Text
		division := -1
		var conference string
		h.DOM.Find("span.panel-heading-normal-text").First().Children().Each(func(i int, s *goquery.Selection) {
			d := parse_division(s.Text())
			if division >= 0 && d >= 0 && division != d {
				logger.Fatalf("Found conflicting divisions in the parsed division list: %d, %d", division, d)
			} else if d >= 0 {
				division = d
			}

			conf, err := regexp.MatchString("conference", strings.ToLower(s.Text()))
			if err != nil {
				logger.Fatalf("Error with coneference search regexp")
			} else if conf {
				conference = s.Text()
			}
		})

		if division < 0 {
			logger.Println("Could not parse a division from the school page")
		}

		if len(conference) == 0 {
			logger.Println("Could not find a conference on the school page")
		}

		school := internal.School{
			Name:       teamName,
			URL:        h.Request.URL.RequestURI(),
			Division:   division,
			Conference: conference,
		}
		db.InsertSchool(school)
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
		athName := e.Text
		schoolNameNode := e.DOM.Parent().SiblingsFiltered("a.underline-hover-white.pl-0.panel-actions")
		schoolNameURL, exists := schoolNameNode.Attr("href")
		if exists == false {
			logger.Fatal("Could not find the href attribute in the athlete title line")
		}
		school, found := db.GetSchoolURL(schoolNameURL)
		if !found {
			logger.Fatal("School not found, we should be able to find it:", schoolNameURL)
		}
		ath := internal.Athlete{
			ID:       school.ID,
			Name:     athName,
			SchoolID: school.ID,
		}
		db.InsertAthlete(ath)
	})
}

func NewMeetCollector(db *database.BacticDB) *colly.Collector {

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

		row_length := resultsRows.First().Children().Length()
		table := make([][]string, tableLength+1)

		table[0] = []string{strings.TrimSpace(strings.Replace(e.ChildText("h3"), "\n", " ", -1))}
		whitespaceReplace := regexp.MustCompile(`\s\s+`)

		athleteURLs := make(map[uint32]string)
		schoolURLs := make([]string, 0, tableLength)
		// Collect into table struct
		resultsRows.Each(func(i int, s *goquery.Selection) {
			table[i+1] = make([]string, row_length)
			s.Children().Each(func(j int, r *goquery.Selection) {

				// strip the athlete id, which we use to identify athletes
				if j == 1 {
					athleteURL, _ := r.Children().Attr("href")
					findID := regexp.MustCompile(`https://www.tfrrs.org/athletes/(\d+)`).FindStringSubmatch(athleteURL)
					if len(findID) < 2 {
						logger.Fatalf("athlete url could not be searched for an id: %s", athleteURL)
					}
					athleteID, err := strconv.Atoi(findID[1])
					if err != nil {
						logger.Print("Unable to convert athlete url into valid ID:", athleteURL)
					}
					athleteURLs[uint32(athleteID)] = athleteURL
					table[i+1][j] = fmt.Sprint(athleteID)
				} else if j == 3 {
                    url, _ := r.Children().Attr("href")
                    schoolURLs = append(schoolURLs, url)
                    table[i+1][j] = strings.TrimSpace(r.Children().Text())
				} else {
					table[i+1][j] = strings.TrimSpace(whitespaceReplace.ReplaceAllString(r.Text(), " "))
				}
			})
		})

		// parse all information from table
		resultTable, eventType := parseResultTable(table, logger)
		athletesToScrape := db.GetMissingAthletes(resultTable)
		schoolsToScrape := db.GetMissingSchools(schoolURLs)

        // visit schools before athletes due to the database relation dependencies
        for _, url := range schoolsToScrape {
            schoolCollector.Visit(url)
        }

		for _, athleteID := range athletesToScrape {
			athleteCollector.Visit(athleteURLs[athleteID])
		}

        // once this is done, we can insert the heat
		db.InsertHeat(eventType, 1, resultTable)
	})
	return meetCollector
}
