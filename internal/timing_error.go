package internal

type TimingError struct {
	Code string `json:"code"`
}

func AssertTimingErrorRequired(obj TimingError) error {
	elements := map[string]interface{}{
		"code": obj.Code,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertResultConstraints checks if the values respects the defined constraints
func AssertTimingErrorConstraints(obj Result) error {
	return nil
}

func (t TimingError) Error() string {
	return t.Code
}
