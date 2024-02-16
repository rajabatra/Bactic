package data

import "bactic/internal"

type Athlete struct {
	Name    string   `json:"name"`
	Schools []uint32 `json:"schools,omitempty"`
	Id      uint32   `json:"id"`
	ZScore  float32  `json:"zscore"`
}

// AssertAthleteRequired checks if the required fields are not zero-ed
func (obj Athlete) AssertRequired() error {
	elements := map[string]interface{}{
		"name":   obj.Name,
		"id":     obj.Id,
		"zscore": obj.ZScore,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}

	return nil
}

// AssertAthleteConstraints checks if the values respects the defined constraints
func (obj Athlete) AssertConstraints() error {
	return nil
}
