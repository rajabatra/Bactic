package data

type AthleteSummary struct {
	Athlete Athlete              `json:"athlete,omitempty"`
	Zscore  float32              `json:"zscore,omitempty"`
	Results map[EventType]Result `json:"results,omitempty"`
}

// AssertAthleteSummaryRequired checks if the required fields are not zero-ed
func (obj AthleteSummary) AssertRequired() error {
	if err := obj.Athlete.AssertRequired(); err != nil {
		return err
	}
	return nil
}

// AssertAthleteSummaryConstraints checks if the values respects the defined constraints
func (obj AthleteSummary) AssertConstraints() error {
	return nil
}
