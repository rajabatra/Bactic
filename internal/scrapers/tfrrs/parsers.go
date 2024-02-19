package tfrrs

import (
	"bactic/internal/data"
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

func extractDivision(region string) string {
	re := regexp.MustCompile(`^di+\s|^naia$`)
	region = strings.TrimSpace(strings.ToLower(region))
	return re.FindString(region)
}

// For parsing integers, handles the case where we have a trailing dot from the regexp capture group
func parseuint32(value string) uint32 {
	// strip the leading dot or colon
	if len(value) > 0 && (value[len(value)-1] == '.' || value[len(value)-1] == ':') {
		value = value[:len(value)-1]
	}

	valInt, err := strconv.Atoi(value)
	if err != nil {
		return 0
	} else {
		return uint32(valInt)
	}
}

func parseTime(t string) (float32, error) {
	if !slices.Contains([]string{"DNF", "DQ", "FS", "DNS", "NT"}, t) {
		time_regexp := regexp.MustCompile(`(\d+:)?(\d+).(\d+)`)
		matches := time_regexp.FindStringSubmatch(t)
		if len(matches) != 4 {
			return 0.0, errors.New("time could not be parsed into expected format string: t")
		}
		minutes := time.Duration(parseuint32(matches[1]))
		seconds := time.Duration(parseuint32(matches[2]))
		frac := time.Duration(parseuint32(matches[3])) * 10

		return float32((minutes*time.Minute + seconds*time.Second + frac*time.Millisecond).Seconds()), nil
	}

	return 0.0, &data.TimingError{Code: t}
}

var titleToEventEnum = map[string]data.EventType{
	"60 meters":         data.T60M,
	"5000 meters":       data.T5000M,
	"5,000 meters":      data.T5000M,
	"100 meters":        data.T100M,
	"200 meters":        data.T200M,
	"400 meters":        data.T400M,
	"800 meters":        data.T800M,
	"1500 meters":       data.T1500M,
	"mile":              data.T1MI,
	"10,000 meters":     data.T10000M,
	"100 hurdles":       data.T100MH,
	"110 hurdles":       data.T110MH,
	"400 hurdles":       data.T400MH,
	"3000 steeplechase": data.T3000MS,
	"3000 meters":       data.T3000M,
	"4 x 100m relay":    data.T4X100M,
	"4 x 100 relay":     data.T4X100M,
	"4 x 400 relay":     data.T4X400M,
	"high jump":         data.HIGH_JUMP,
	"pole vault":        data.VAULT,
	"long jump":         data.LONG_JUMP,
	"triple jump":       data.TRIPLE_JUMP,
	"shot put":          data.SHOT,
	"weight throw":      data.WEIGHT_THROW,
	"discus":            data.DISCUS,
	"hammer":            data.HAMMER,
	"javelin":           data.JAV,
	"decathlon":         data.DEC,
	"heptathlon":        data.HEPT,
}

// Given an xc table header, return whether or not it is a summary
func parseXCEventType(eventTitle string) (data.EventType, error) {
	eventTitle = strings.ToLower(eventTitle)
	if strings.Contains(eventTitle, "8k") {
		return data.XC8K, nil
	} else if strings.Contains(eventTitle, "10k") {
		return data.XC10K, nil
	} else if strings.Contains(eventTitle, "6k") {
		return data.XC6K, nil
	} else {
		return 0, errors.New("event table could not be parsed for type")
	}
}

// Given an event title, return the enumerated event
func parseEvent(eventTitle string) (data.EventType, error) {
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

func parseResultTable(resultTable [][]htmlElement, logger *log.Logger, eventType data.EventType) (results []data.Result, linkIds []uint32, schoolUrls []string) {
	ret := make([]data.Result, 0, len(resultTable))
	athleteIDs := make([]uint32, 0, len(resultTable))
	schoolURLs := make([]string, 0, len(resultTable))

	for _, row := range resultTable {
		result, athleteID, schoolURL, err := parseIndividualResultClass[eventType](row)
		if err != nil {
			logger.Printf("Unable to parse event row %v due to error: %v. Ignoring", row, err)
		} else {
			ret = append(ret, result)
			schoolURLs = append(schoolURLs, schoolURL)
			athleteIDs = append(athleteIDs, athleteID)
		}
	}
	return ret, athleteIDs, schoolURLs
}

func parseXCResult(row []htmlElement) (result data.Result, athleteID uint32, schoolURL string, errRet error) {
	place, err := strconv.Atoi(row[0].text)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	athleteID, err = parseAthleteIDFromURL(row[1].link)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	time, err := parseTime(row[5].text)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	return data.Result{
		Place:    uint32(place),
		Quantity: time,
	}, athleteID, row[3].link, nil
}

func parseDistanceResult(row []htmlElement) (result data.Result, athleteID uint32, schoolURL string, err error) {
	if len(row) < 5 {
		return data.Result{}, 0, "", fmt.Errorf("distance result row %v is less than the correct length of 4", row)
	}
	place, err := strconv.Atoi(row[0].text)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	time, err := parseTime(row[4].text)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	athleteID, err = parseAthleteIDFromURL(row[1].link)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	return data.Result{
		Quantity: time,
		Place:    uint32(place),
	}, athleteID, row[3].link, nil
}

func parseSprintsResult(row []htmlElement) (result data.Result, athleteID uint32, schoolURL string, err error) {
	if len(row) < 5 {
		return data.Result{}, 0, "", fmt.Errorf("distance result row %v is less than the correct length of 4", row)
	}

	place, err := strconv.Atoi(row[0].text)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	time, err := parseTime(row[4].text)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	athleteID, err = parseAthleteIDFromURL(row[1].link)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	return data.Result{
		Quantity: time,
		Place:    uint32(place),
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
func parseRelayResult(row []htmlElement) (result data.Result, athleteID uint32, schoolURL string, err error) {
	if len(row) < 5 {
		return data.Result{}, 0, "", fmt.Errorf("relay result row %v is less than the correct length of 5", row)
	}

	if len(row) < 4 {
		return data.Result{}, 0, "", fmt.Errorf("the row was less than length 4")
	}

	time, err := parseTime(row[3].text)
	if err != nil {
		return data.Result{}, 0, "", err
	}

	err = nil
	members := make([]uint32, 4)
	for i, m := range strings.Split(row[2].link, ", ") {
		members[i], err = parseAthleteIDFromURL(m)
		if err != nil {
			return data.Result{}, 0, "", err
		}
	}

	return data.Result{
		Quantity: time,
		Members:  members,
	}, 0, row[0].link, nil
}

func notImplementedResult(row []htmlElement) (data.Result, uint32, string, error) {
	return data.Result{}, 0, "", errors.New("this result parser not yet implemented")
}

var parseIndividualResultClass = map[data.EventType](func([]htmlElement) (result data.Result, athleteID uint32, schoolURL string, err error)){
	data.T60M:         parseSprintsResult,
	data.T5000M:       parseDistanceResult,
	data.T100M:        parseSprintsResult,
	data.T200M:        parseSprintsResult,
	data.T400M:        parseSprintsResult,
	data.T800M:        parseDistanceResult,
	data.T1500M:       parseDistanceResult,
	data.T1MI:         parseDistanceResult,
	data.T10000M:      parseDistanceResult,
	data.T100MH:       notImplementedResult,
	data.T110MH:       parseSprintsResult,
	data.T400MH:       parseSprintsResult,
	data.T3000MS:      parseDistanceResult,
	data.T3000M:       parseDistanceResult,
	data.T4X100M:      parseRelayResult,
	data.T4X400M:      parseRelayResult,
	data.HIGH_JUMP:    notImplementedResult,
	data.VAULT:        notImplementedResult,
	data.LONG_JUMP:    notImplementedResult,
	data.TRIPLE_JUMP:  notImplementedResult,
	data.SHOT:         notImplementedResult,
	data.WEIGHT_THROW: notImplementedResult,
	data.DISCUS:       notImplementedResult,
	data.HAMMER:       notImplementedResult,
	data.JAV:          notImplementedResult,
	data.DEC:          notImplementedResult,
	data.HEPT:         notImplementedResult,
	data.XC10K:        parseXCResult,
	data.XC8K:         parseXCResult,
	data.XC6K:         parseXCResult,
}
