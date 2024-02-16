package data

import (
	"bactic/internal"
	"time"
)

type Meet struct {
	Date   time.Time `json:"date"`
	Name   string    `json:"meet_id"`
	Season Season    `json:"season"`
	Id     uint32    `json:"id"`
}

// AssertHeatRequired checks if the required fields are not zero-ed
func (obj Meet) AssertRequired() error {
	elements := map[string]interface{}{
		"id":      obj.Id,
		"meet_id": obj.Name,
		"season":  obj.Season,
		"date":    obj.Date,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}
	return nil
}

// AssertHeatConstraints checks if the values respects the defined constraints
func (obj Meet) AssertConstraints() error {
	return nil
}
