package data

import "bactic/internal"

type StatsTimeseriesGet200ResponseDataInner struct {
	Mean float32 `json:"mean"`

	Variance float32 `json:"variance"`
}

// AssertStatsTimeseriesGet200ResponseDataInnerRequired checks if the required fields are not zero-ed
func (obj StatsTimeseriesGet200ResponseDataInner) AssertRequired() error {
	elements := map[string]interface{}{
		"mean":     obj.Mean,
		"variance": obj.Variance,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}

	return nil
}

// AssertStatsTimeseriesGet200ResponseDataInnerConstraints checks if the values respects the defined constraints
func (obj StatsTimeseriesGet200ResponseDataInner) AssertConstraints() error {
	return nil
}
