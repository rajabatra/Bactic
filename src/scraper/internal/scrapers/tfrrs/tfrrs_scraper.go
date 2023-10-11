package scrapers

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
	switch region {
	case "DIII":
		return internal.DIII
	case "DII":
		return internal.DII
	case "DI":
		return internal.DI
	case "NAIA":
		return internal.NAIA
	default:
		return -1
	}
}

// For parsing integers, handles the case where we have a trailing dot from the regexp capture group
func parseInt64(value string) int64 {

	// strip the leading dot or colon
	if len(value) > 0 && value[len(value)-1] == '.' || value[len(value)-1] == ':' {
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
		logger.Print("Unable to parse event title", err)
		return []internal.Result{}, 0
	}

	ret := make([]internal.Result, 0, len(resultTable)-1)

	for i := 1; i <= len(resultTable); i++ {
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
	place, err := strconv.Atoi(row[0])
	if err != nil {
		return internal.Result{}, fmt.Errorf("Unable to parse place string %e", err)
	}

	athleteID, err := strconv.Atoi(row[1])
	if err != nil {
		return internal.Result{}, fmt.Errorf("Unable to parse athlete id string %e", err)
	}

	time, err := parseTime(row[3])
	if err != nil {
		return internal.Result{}, err
	}

	return internal.Result{
		AthleteID: uint32(athleteID),
		Place:     uint8(place),
		Quantity:  time,
	}, nil
}

func parseSprintsResult(row []string) (internal.Result, error) {
	return internal.Result{}, errors.New("Spring result parser not yet implemented")
}

func notImplementedResult(row []string) (internal.Result, error) {
	return internal.Result{}, errors.New("Spring result parser not yet implemented")
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

func NewTFRRSCollector(db *database.BacticDB, logger *log.Logger) TFRRSParser {
	schoolCollector := colly.NewCollector()
	athleteCollector := colly.NewCollector()
	meetCollector := colly.NewCollector()
	rootCollector := colly.NewCollector()

    // setup collector
    setupMeetCollector(meetCollector, athleteCollector, schoolCollector, db, logger)

	return rootCollector
}

func setupSchoolCollector(schoolCollector *colly.Collector, db *database.BacticDB, logger *log.Logger) {
}

func setupAthleteCollector(athleteCollector *colly.Collector, schoolCollector *colly.Collector, db *database.BacticDB, logger *log.Logger) {
    setupSchoolCollector(schoolCollector, db, logger)

    athleteCollector.OnRequest(func (r *colly.Request ) {
        logger.Println("Athlete collector visiting athlete", r.URL)
    })

    athleteCollector.OnHTML("h3.panel-title.large-title", func (e *colly.HTMLElement) {
        // what info do we need?
        athName := e.Text
        schoolNameNode := e.DOM.Parent().Siblings().SiblingsMatcher("a.underline-hover-white.pl-0.panel-actions")
        schoolNameURL, exists := schoolNameNode.Attr("href")
        if exists == false {
        }
        

    })
}

func setupMeetCollector(meetCollector *colly.Collector, athleteCollector *colly.Collector, schoolCollector *colly.Collector, db *database.BacticDB, logger *log.Logger) {
    //recursively setup athlete collector
    setupAthleteCollector(athleteCollector, schoolCollector, db, logger)

	meetCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Visiting meet", r.URL)
	})

	meetCollector.OnHTML("div.row", func(e *colly.HTMLElement) {
		resultsRows := e.DOM.Find("tbody>tr")
		table_length := resultsRows.Length()
		if table_length == 0 {
			return
		}

		row_length := resultsRows.First().Children().Length()
		table := make([][]string, table_length+1)

		table[0] = []string{strings.TrimSpace(strings.Replace(e.ChildText("h3"), "\n", " ", -1))}
		whitespaceReplace := regexp.MustCompile(`\s\s+`)

		athleteURLs := make(map[uint32]string)
		// Collect into table struct
		resultsRows.Each(func(i int, s *goquery.Selection) {
			table[i+1] = make([]string, row_length)
			s.Children().Each(func(j int, r *goquery.Selection) {

				// strip the athlete id, which we use to identify athletes
				if j == 1 {
					athleteURL, _ := r.Children().Attr("href")
					athleteRegexp := regexp.MustCompile(`https://www.tfrrs.org/athletes/(\d+)`)
					athleteID, err := strconv.Atoi(athleteRegexp.FindStringSubmatch(athleteURL)[1])
                    if err != nil {
                        logger.Print("Unable to convert athlete url into valid ID:", athleteURL)
                    }
					athleteURLs[uint32(athleteID)] = athleteURL
					table[i+1][j] = string(athleteID)
				} else {
					table[i+1][j] = strings.TrimSpace(whitespaceReplace.ReplaceAllString(r.Text(), " "))
				}
			})
		})

		// parse all information from table
		resultTable, eventType := parseResultTable(table, logger)
		athletesToScrape := db.GetMissingAthletes(resultTable)
		for _, athleteID := range athletesToScrape {
			meetCollector.Visit(athleteURLs[athleteID])
		}

	})
}
