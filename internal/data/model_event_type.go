package data

import (
	"fmt"
	"log"
)

type EventType uint32

// List of EventType
const (
	T60M EventType = iota << 1
	T100M
	T200M
	T400M
	T800M
	T1500M
	T1MI
	T3000M
	T3000MS
	T2MI
	T5000M
	T10000M
	T4X100M
	T4X400M
	HIGH_JUMP
	VAULT
	LONG_JUMP
	TRIPLE_JUMP
	SHOT
	WEIGHT_THROW
	DISCUS
	HAMMER
	JAV
	DEC
	HEPT
	T100MH
	T110MH
	T400MH
	XC6K
	XC8K
	XC10K
)

func (ev EventType) AssertValid() bool {
	return ev >= T60M && ev <= XC10K
}

var eventTypeToString = map[EventType]string{
	T60M:         "60 Meters",
	T100M:        "100 Meters",
	T200M:        "200 Meters",
	T400M:        "400 Meters",
	T800M:        "800 Meters",
	T1500M:       "1500 Meters",
	T1MI:         "1 Mile",
	T3000M:       "3000 Meters",
	T3000MS:      "3000 Meter Steeplechase",
	T2MI:         "2 Mile",
	T5000M:       "5,000 Meters",
	T10000M:      "10,000 Meters",
	T4X100M:      "4x100 Meter Relay",
	T4X400M:      "4x400 Meter Relay",
	HIGH_JUMP:    "High Jump",
	VAULT:        "Pole Vault",
	LONG_JUMP:    "Long Jump",
	TRIPLE_JUMP:  "Triple Jump",
	SHOT:         "Shot Put",
	WEIGHT_THROW: "Weight Throw",
	DISCUS:       "Discus",
	HAMMER:       "Hammer Throw",
	JAV:          "Javelin",
	DEC:          "Decathlon",
	HEPT:         "Heptathlon",
	T100MH:       "100 Meter Hurdles",
	T110MH:       "110 Meter Hurdles",
	T400MH:       "400 Meter Hurdles",
	XC6K:         "Cross Country 6K",
	XC8K:         "Cross Country 8K",
	XC10K:        "Cross Country 10K",
}

var titleToEventEnum = map[string]EventType{
	"60 meters":         T60M,
	"5000 meters":       T5000M,
	"5,000 meters":      T5000M,
	"100 meters":        T100M,
	"200 meters":        T200M,
	"400 meters":        T400M,
	"800 meters":        T800M,
	"1500 meters":       T1500M,
	"mile":              T1MI,
	"10,000 meters":     T10000M,
	"100 hurdles":       T100MH,
	"110 hurdles":       T110MH,
	"400 hurdles":       T400MH,
	"3000 steeplechase": T3000MS,
	"3000 meters":       T3000M,
	"4 x 100m relay":    T4X100M,
	"4 x 100 relay":     T4X100M,
	"4 x 400 relay":     T4X400M,
	"high jump":         HIGH_JUMP,
	"pole vault":        VAULT,
	"long jump":         LONG_JUMP,
	"triple jump":       TRIPLE_JUMP,
	"shot put":          SHOT,
	"weight throw":      WEIGHT_THROW,
	"discus":            DISCUS,
	"hammer":            HAMMER,
	"javelin":           JAV,
	"decathlon":         DEC,
	"heptathlon":        HEPT,
}

func (ev EventType) String() string {
	if !ev.AssertValid() {
		log.Fatalf("EventType %d is not valid", uint32(ev))
	}
	return eventTypeToString[ev]
}

// NewEventTypeFromValue returns a pointer to a valid EventType
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEventTypeFromValue(v string) (EventType, error) {
	ev, found := titleToEventEnum[v]
	if found {
		return ev, nil
	} else {
		return 0, fmt.Errorf("invalid value '%v' for EventType: valid values are %v", v, titleToEventEnum)
	}
}

// AssertEventTypeRequired checks if the required fields are not zero-ed
func (obj EventType) AssertRequired() error {
	return nil
}

// AssertEventTypeConstraints checks if the values respects the defined constraints
func (obj EventType) AssertConstraints() error {
	return nil
}
