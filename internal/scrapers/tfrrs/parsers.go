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
			return 0.0, errors.New("Time could not be parsed into expected format string: t")
		}
		minutes := time.Duration(parseInt64(matches[1]))
		seconds := time.Duration(parseInt64(matches[2]))
		frac := time.Duration(parseInt64(matches[3])) * 10

		return float32((minutes*time.Minute + seconds*time.Second + frac*time.Millisecond).Seconds()), nil
	}
}

var titleToEventEnum = map[string]internal.EventType{
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

// Given an xc table header, return whether or not it is a summary
func parseXCEventType(eventTitle string) (internal.EventType, error) {
	eventTitle = strings.ToLower(eventTitle)
	if strings.Contains("8k", eventTitle) {
		return internal.XC_8K, nil
	} else if strings.Contains("10k", eventTitle) {
		return internal.XC_10K, nil
	} else if strings.Contains("6k", eventTitle) {
		return internal.XC_6K, nil
	} else {
		return 0, errors.New("Event table could not be parsed for type")
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
	if exists != true {
		return 0, fmt.Errorf("The event title %s, which converts to key %s could not be mapped to an event", eventTitle, eventParsed)
	}

	return event_type, nil
}

type ParseResult struct {
	result internal.Result
	err    error
}

func parseResultTable(resultTable [][][]string, logger *log.Logger, eventType internal.EventType) ([]internal.Result, []uint32, []string) {
	ret := make([]internal.Result, 0, len(resultTable))
	athleteIDs := make([]uint32, 0, len(resultTable))
	schoolURLs := make([]string, 0, len(resultTable))

	for _, row := range resultTable {
		result, athleteID, schoolURL, err := parseResultClass[eventType](row)
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

func parseXCResult(row [][]string) (internal.Result, uint32, string, error) {
	place, err := strconv.Atoi(row[0][0])
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	athleteID, err := parseAthleteIDFromURL(row[1][1])
	if err != nil {
		return internal.Result{}, 0, "", nil
	}

	time, err := parseTime(row[5][0])
	if err != nil {
		return internal.Result{}, 0, "", nil
	}

	return internal.Result{
		Place:    place,
		Quantity: time,
	}, athleteID, row[3][1], nil
}
func parseDistanceResult(row [][]string) (internal.Result, uint32, string, error) {
	if len(row) < 5 {
		return internal.Result{}, 0, "", fmt.Errorf("Distance result row %v is less than the correct length of 4", row)
	}
	var (
		err   error
		place int
	)

	if len(row[0][0]) > 0 {
		place, err = strconv.Atoi(row[0][0])
		if err != nil {
			return internal.Result{}, 0, "", err
		}
	}

	time, err := parseTime(row[4][0])
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	athleteID, err := parseAthleteIDFromURL(row[1][1])
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	return internal.Result{
		Quantity: time,
		Place:    place,
	}, athleteID, row[3][1], nil
}

func parseSprintsResult(row [][]string) (internal.Result, uint32, string, error) {
	if len(row) < 5 {
		return internal.Result{}, 0, "", fmt.Errorf("Distance result row %v is less than the correct length of 4", row)
	}

	var (
		place int
		err   error
	)
	if len(row[0][0]) > 0 {
		place, err = strconv.Atoi(row[0][0])
		if err != nil {
			return internal.Result{}, 0, "", err
		}
	}

	time, err := parseTime(row[4][0])
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	athleteID, err := parseAthleteIDFromURL(row[1][1])
	if err != nil {
		return internal.Result{}, 0, "", err
	}

	return internal.Result{
		Quantity: time,
		Place:    place,
	}, athleteID, row[3][1], nil
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

func notImplementedResult(row [][]string) (internal.Result, uint32, string, error) {
	return internal.Result{}, 0, "", errors.New("This result parser not yet implemented")
}

var parseResultClass = map[internal.EventType](func([][]string) (internal.Result, uint32, string, error)){
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
	internal.XC_10K:      parseXCResult,
	internal.XC_8K:       parseXCResult,
	internal.XC_6K:       parseXCResult,
}
