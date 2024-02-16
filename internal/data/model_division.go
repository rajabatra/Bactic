package data

import (
	"fmt"
)

type Division uint32

// List of EventType
const (
	DI Division = iota << 1
	DII
	DIII
	NAIA
)

// AllowedEventTypeEnumValues is all the allowed values of EventType enum
func (div Division) AssertValid() bool {
	return div >= DI && div <= NAIA
}

var divisionToString = map[Division]string{
	DI:   "NCAA DI",
	DII:  "NCAA DII",
	DIII: "NCAA DIII",
	NAIA: "NAIA",
}

func (div Division) String() string {
	if !div.AssertValid() {
		return "Tried to print a division with invalid value"
	}
	return divisionToString[div]
}

var titleToDivisionEnum = map[string]Division{
	"di":   DI,
	"dii":  DII,
	"diii": DIII,
	"naia": NAIA,
}

// NewEventTypeFromValue returns a pointer to a valid EventType
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewDivisionFromValue(v string) (Division, error) {
	div, found := titleToDivisionEnum[v]
	if found {
		return div, nil
	} else {
		return 0, fmt.Errorf("invalid value '%s' for Division: valid values are %v", v, titleToEventEnum)
	}
}

// AssertEventTypeRequired checks if the required fields are not zero-ed
func (obj Division) AssertRequired() error {
	return nil
}

// AssertEventTypeConstraints checks if the values respects the defined constraints
func (obj Division) AssertConstraints() error {
	return nil
}
