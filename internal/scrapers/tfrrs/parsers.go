package tfrrs

import (
	"bactic/internal"
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Parse a date string in the tfrrs website format
func parseMeetDate(date string) (time.Time, error) {
	replace_daterange := regexp.MustCompile(`-\d{1,2}`)
	date = string(replace_daterange.ReplaceAll([]byte(date), []byte("")))
	return time.Parse("January 2, 2006", date)
}

func parseDivision(region string) int {
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
			return 0.0, errors.New("time could not be parsed into expected format string: t")
		}
		minutes := time.Duration(parseInt64(matches[1]))
		seconds := time.Duration(parseInt64(matches[2]))
		frac := time.Duration(parseInt64(matches[3])) * 10

		return float32((minutes*time.Minute + seconds*time.Second + frac*time.Millisecond).Seconds()), nil
	}
}

var titleToEventEnum = map[string]internal.EventType{
	"60 meters":         internal.T60M,
	"5000 meters":       internal.T5000M,
	"5,000 meters":      internal.T5000M,
	"100 meters":        internal.T100M,
	"200 meters":        internal.T200M,
	"400 meters":        internal.T400M,
	"800 meters":        internal.T800M,
	"1500 meters":       internal.T1500M,
	"mile":              internal.T1MILE,
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
	"weight throw":      internal.WEIGHT_THROW,
	"discus":            internal.DISCUS,
	"hammer":            internal.HAMMER,
	"javelin":           internal.JAV,
	"decathlon":         internal.DEC,
	"heptathlon":        internal.HEPT,
	"100 hurdles":       internal.T100H,
}

// Given an xc table header, return whether or not it is a summary
func parseXCEventType(eventTitle string) (internal.EventType, error) {
	eventTitle = strings.ToLower(eventTitle)
	if strings.Contains(eventTitle, "8k") {
		return internal.XC_8K, nil
	} else if strings.Contains(eventTitle, "10k") {
		return internal.XC_10K, nil
	} else if strings.Contains(eventTitle, "6k") {
		return internal.XC_6K, nil
	} else {
		return 0, errors.New("event table could not be parsed for type")
	}
}

// Given an event title, return the enumerated event
func parseEvent(eventTitle string) (internal.EventType, error) {
	sexRe := regexp.MustCompile(`Men's|Women's`)
	trailingRe := regexp.MustCompile(`(Preliminaries|Finals)?(Heat {\d+})?`)
	eventParsed := sexRe.ReplaceAllString(eventTitle, "")
	eventParsed = trailingRe.ReplaceAllString(eventParsed, "")
	eventParsed = strings.TrimSpace(eventParsed)
	eventParsed = strings.ToLower(eventParsed)

	event_type, exists := titleToEventEnum[eventParsed]
	if !exists {
		return 0, fmt.Errorf("the event title %s, which converts to key %s could not be mapped to an event", eventTitle, eventParsed)
	}

	return event_type, nil
}

type htmlElement struct {
	text string
	link string
}

func parseResultTable(resultTable [][]htmlElement, logger *log.Logger, eventType internal.EventType) (results []internal.Result, linkids []uint32, schoolurls []string) {
	ret := make([]internal.Result, 0, len(resultTable))
	athleteIDs := make([]uint32, 0, len(resultTable))
	schoolURLs := make([]string, 0, len(resultTable))

	for _, row := range resultTable {
		result, athleteID, schoolURL, err := parseIndividualResultClass[eventType](row)
		if err != nil {
			logger.Printf("Unable to parse event row due to error: %v. Ignoring", err)
		} else {
			ret = append(ret, result)
			schoolURLs = append(schoolURLs, schoolURL)
			athleteIDs = append(athleteIDs, athleteID)
		}
	}
	return ret, athleteIDs, schoolURLs
}

func parseXCResult(row []htmlElement) (result internal.Result, athleteID uint32, schoolURL string, errRet error) {
	place, err := strconv.Atoi(row[0].text)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	athleteID, err = parseAthleteIDFromURL(row[1].link)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	time, err := parseTime(row[5].text)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	return internal.Result{
		Place:    place,
		Quantity: time,
	}, athleteID, row[3].link, nil
}
func parseDistanceResult(row []htmlElement) (result internal.Result, athleteID uint32, schoolURL string, err error) {
	if len(row) < 5 {
		return internal.Result{}, 0, "", fmt.Errorf("distance result row %v is less than the correct length of 4", row)
	}
	place, err := strconv.Atoi(row[0].text)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	time, err := parseTime(row[4].text)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	athleteID, err = parseAthleteIDFromURL(row[1].link)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	return internal.Result{
		Quantity: time,
		Place:    place,
	}, athleteID, row[3].link, nil
}

func parseSprintsResult(row []htmlElement) (result internal.Result, athleteID uint32, schoolURL string, err error) {
	if len(row) < 5 {
		return internal.Result{}, 0, "", fmt.Errorf("distance result row %v is less than the correct length of 4", row)
	}

	place, err := strconv.Atoi(row[0].text)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	time, err := parseTime(row[4].text)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	athleteID, err = parseAthleteIDFromURL(row[1].link)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	return internal.Result{
		Quantity: time,
		Place:    place,
	}, athleteID, row[3].link, nil
}

func parseAthleteIDFromURL(athleteURL string) (uint32, error) {
	// if this is not a tfrrs id, return the nil reference
	if len(athleteURL) == 0 || !strings.HasPrefix(athleteURL, "https://www.tfrrs.org/athletes/") {
		return 0, nil // Return the nil reference id
	}
	findTFRRSID := regexp.MustCompile(`https://www.tfrrs.org/athletes/(\d+)`).FindStringSubmatch(athleteURL)
	if len(findTFRRSID) < 2 {
		return 0, fmt.Errorf("athlete url could not be searched for an id: %s", athleteURL)
	}
	athleteID, err := strconv.Atoi(findTFRRSID[1])
	if err != nil {
		return 0, fmt.Errorf("unable to convert athlete url into valid ID: %s", athleteURL)
	}
	return uint32(athleteID), nil
}

// TODO: parse relay results has not been tested
func parseRelayResult(row []htmlElement) (result internal.Result, athleteID uint32, schoolURL string, err error) {
	if len(row) < 5 {
		return internal.Result{}, 0, "", fmt.Errorf("relay result row %v is less than the correct length of 5", row)
	}

	if len(row) < 4 {
		return internal.Result{}, 0, "", fmt.Errorf("the row was less than length 4")
	}

	time, err := parseTime(row[3].text)
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	err = nil
	members := make([]uint32, 4)
	for i, m := range strings.Split(row[2].link, ", ") {
		members[i], err = parseAthleteIDFromURL(m)
		if err != nil {
			return internal.Result{}, 0, "", err
		}
	}

	return internal.Result{
		Quantity: time,
		Members:  members,
	}, 0, row[0].link, nil
}

func notImplementedResult(row []htmlElement) (internal.Result, uint32, string, error) {
	return internal.Result{}, 0, "", errors.New("this result parser not yet implemented")
}

var parseIndividualResultClass = map[internal.EventType](func([]htmlElement) (result internal.Result, athleteID uint32, schoolURL string, err error)){
	internal.T60M:        parseSprintsResult,
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
	internal.T4X100:      parseRelayResult,
	internal.T4X400:      parseRelayResult,
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
	internal.XC_10K:      parseXCResult,
	internal.XC_8K:       parseXCResult,
	internal.XC_6K:       parseXCResult,
}
