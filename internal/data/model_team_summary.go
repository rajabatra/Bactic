package data

import "bactic/internal"

type TeamSummary struct {
	Name     string    `json:"name"`
	Athletes []Athlete `json:"athletes,omitempty"`
	Id       uint32    `json:"id"`
	Zscore   float32   `json:"zscore,omitempty"`
}

// AssertTeamSummaryRequired checks if the required fields are not zero-ed
func AssertTeamSummaryRequired(obj TeamSummary) error {
	elements := map[string]interface{}{
		"name": obj.Name,
		"id":   obj.Id,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}

	for _, el := range obj.Athletes {
		if err := el.AssertRequired(); err != nil {
			return err
		}
	}
	return nil
}

// AssertTeamSummaryConstraints checks if the values respects the defined constraints
func AssertTeamSummaryConstraints(obj TeamSummary) error {
	return nil
}
