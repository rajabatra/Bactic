package data

import "bactic/internal"

type Heat struct {
	Type   *interface{} `json:"type"`
	Id     uint32       `json:"id"`
	MeetId uint32       `json:"meet_id"`
}

// AssertHeatRequired checks if the required fields are not zero-ed
func (obj Heat) AssertRequired() error {
	elements := map[string]interface{}{
		"id":      obj.Id,
		"meet_id": obj.MeetId,
		"type":    obj.Type,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}

	return nil
}

// AssertHeatConstraints checks if the values respects the defined constraints
func (obj Heat) AssertConstraints() error {
	return nil
}
