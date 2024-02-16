package data

import "bactic/internal"

type Result struct {
	Team      string   `json:"team,omitempty"`
	Members   []uint32 `json:"members,omitempty"`
	Id        uint32   `json:"id"`
	HeatId    uint32   `json:"heat_id"`
	AthleteId uint32   `json:"athlete_id"`
	Place     int32    `json:"place"`
	Quantity  float32  `json:"quantity"`
	WindMs    float32  `json:"wind_ms,omitempty"`
	Stage     int32    `json:"stage,omitempty"`
}

// AssertResultRequired checks if the required fields are not zero-ed
func (obj Result) AssertRequired() error {
	elements := map[string]interface{}{
		"id":         obj.Id,
		"heat_id":    obj.HeatId,
		"athlete_id": obj.AthleteId,
		"place":      obj.Place,
		"quantity":   obj.Quantity,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}

	return nil
}

// AssertResultConstraints checks if the values respects the defined constraints
func (obj Result) AssertConstraints() error {
	return nil
}
