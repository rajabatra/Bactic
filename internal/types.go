package internal

import (
	"fmt"
	"time"
)

// Division types
const (
	DIII = 3
	DII  = 2
	DI   = 1
	NAIA = 0
)

var divisionToStr = map[int]string{
	DIII: "DIII",
	DII:  "DII",
	DI:   "DI",
	NAIA: "NAIA",
}

// Event types
const (
	T5000M      = 0
	T100M       = 1
	T200M       = 2
	T400M       = 3
	T800M       = 4
	T1500M      = 5
	T10000M     = 6
	T110H       = 7
	T400H       = 8
	T3000S      = 9
	T3000M      = 20
	T4X100      = 10
	T4X400      = 11
	HIGH_JUMP   = 12
	VAULT       = 13
	LONG_JUMP   = 14
	TRIPLE_JUMP = 15
	SHOT        = 16
	DISCUS      = 17
	HAMMER      = 18
	JAV         = 19
	DEC         = 21
	HEPT        = 22
	T100H       = 23
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
}

// Event stages
const (
	PRELIM = 0
	FINAL  = 1
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
	Type      uint8
	Place     uint8
	// Either time in seconds or meters for distance respective of the event type
	Quantity float32
	WindMS   float32
	Stage    uint8
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
	ID   uint32
	Name string
	Date time.Time
}

type Athlete struct {
	ID      uint32
	Name    string
	Schools []uint32 // athelete can be part of multiple schools
}
