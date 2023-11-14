package internal

import (
	"fmt"
	"time"
)

// Division types
const (
	DIII = iota
	DII  = iota
	DI   = iota
	NAIA = iota
)

// Seasons
const (
	XC      = iota
	INDOOR  = iota
	OUTDOOR = iota
)

// Seasons
const (
	XC      = 0
	INDOOR  = 1
	OUTDOOR = 2
)

var divisionToStr = map[int]string{
	DIII: "DIII",
	DII:  "DII",
	DI:   "DI",
	NAIA: "NAIA",
}

// Event types (XC and TF)
const (
	T5000M      = iota
	T100M       = iota
	T200M       = iota
	T400M       = iota
	T800M       = iota
	T1500M      = iota
	T10000M     = iota
	T110H       = iota
	T400H       = iota
	T3000S      = iota
	T3000M      = iota
	T4X100      = iota
	T4X400      = iota
	HIGH_JUMP   = iota
	VAULT       = iota
	LONG_JUMP   = iota
	TRIPLE_JUMP = iota
	SHOT        = iota
	DISCUS      = iota
	HAMMER      = iota
	JAV         = iota
	DEC         = iota
	HEPT        = iota
	T100H       = iota
	XC_6K       = iota
	XC_8K       = iota
	XC_10K      = iota
)

var eventToStr = map[int]string{
	T5000M:      "5000m",
	T100M:       "100m",
	T200M:       "200m",
	T400M:       "400m",
	T800M:       "800m",
	T1500M:      "1500m",
	T10000M:     "10000m",
	T110H:       "110 Hurdles",
	T400H:       "400 Hurdles",
	T3000S:      "3000 Steeplechase",
	T3000M:      "3000m",
	T4X100:      "4x100 Relay",
	T4X400:      "4x400",
	HIGH_JUMP:   "High Jump",
	VAULT:       "Vault",
	LONG_JUMP:   "Long Jump",
	TRIPLE_JUMP: "Triple Jump",
	SHOT:        "Shot Put",
	DISCUS:      "Discus",
	HAMMER:      "Hammer",
	JAV:         "Javelin",
	DEC:         "Decathlon",
	HEPT:        "Heptathlon ",
	T100H:       "100 Hurdles",
	XC_10K:      "XC 10K",
	XC_8K:       "XC 8K",
	XC_6K:       "XC 6K",
}

// Event stages
const (
	PRELIM = iota
	FINAL  = iota
)

var stageToString = map[int]string{
	PRELIM: "Prelim",
	FINAL:  "Final",
}

// Timing errors
type TimingError struct {
	Name string
}

func (e *TimingError) Error() string {
	return fmt.Sprintf("Time zero due to timing condition %s", e.Name)
}

// Event Result
type Result struct {
	ID        uint32
	HeatID    uint32
	AthleteID uint32
	Type      int
	Place     int
	// Either time in seconds or meters for distance respective of the event type
	Quantity float32
	WindMS   float32
	Stage    int
	Team     string
	Members  []uint32
}

// TODO: implement
func (m *Result) String() string {
	return "TODO: implement result String()"
}

type Heat struct {
	ID     uint32
	Type   int
	MeetID uint32
}

type School struct {
	ID       uint32
	Name     string
	Division int
	URL      string
	Leagues  []string
}

type Meet struct {
	ID     uint32
	Name   string
	Season int
	Date   time.Time
}

type Athlete struct {
	ID      uint32
	Name    string
	Schools []uint32 // athelete can be part of multiple schools
}
