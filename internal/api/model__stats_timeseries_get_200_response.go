/*
 * Bactic
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.1.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api




type StatsTimeseriesGet200Response struct {

	Data []StatsTimeseriesGet200ResponseDataInner `json:"data"`

	Axis []string `json:"axis"`
}

// AssertStatsTimeseriesGet200ResponseRequired checks if the required fields are not zero-ed
func AssertStatsTimeseriesGet200ResponseRequired(obj StatsTimeseriesGet200Response) error {
	elements := map[string]interface{}{
		"data": obj.Data,
		"axis": obj.Axis,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.Data {
		if err := AssertStatsTimeseriesGet200ResponseDataInnerRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertStatsTimeseriesGet200ResponseConstraints checks if the values respects the defined constraints
func AssertStatsTimeseriesGet200ResponseConstraints(obj StatsTimeseriesGet200Response) error {
	return nil
}
