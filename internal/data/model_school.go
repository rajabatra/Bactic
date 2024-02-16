package data

import "bactic/internal"

type School struct {
	Name     string   `json:"name"`
	Division Division `json:"division,omitempty"`
	Url      string   `json:"url,omitempty"`
	Leagues  []string `json:"leagues,omitempty"`
	Id       uint32   `json:"id"`
}

// AssertSchoolRequired checks if the required fields are not zero-ed
func AssertSchoolRequired(obj School) error {
	elements := map[string]interface{}{
		"id":   obj.Id,
		"name": obj.Name,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}

	return nil
}

// AssertSchoolConstraints checks if the values respects the defined constraints
func AssertSchoolConstraints(obj School) error {
	return nil
}
