package data

type TimingError struct {
	Code string `json:"code"`
}

func (t TimingError) Error() string {
	return "timing error: " + t.Code
}
