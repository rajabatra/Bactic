package data

import "bactic/internal"

type StatsTimeseriesGet200Response struct {
	Data []StatsTimeseriesGet200ResponseDataInner `json:"data"`

	Axis []string `json:"axis"`
}

// AssertStatsTimeseriesGet200ResponseRequired checks if the required fields are not zero-ed
func (obj StatsTimeseriesGet200Response) Assertrequired() error {
	elements := map[string]interface{}{
		"data": obj.Data,
		"axis": obj.Axis,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}

	for _, el := range obj.Data {
		if err := el.AssertRequired(); err != nil {
			return err
		}
	}
	return nil
}

// AssertStatsTimeseriesGet200ResponseConstraints checks if the values respects the defined constraints
func (obj StatsTimeseriesGet200Response) AssertConstraints() error {
	return nil
}
