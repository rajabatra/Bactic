package stats

// An interface for computable and query-capable stats
type Stat interface {
	Compute(string) error
	Query(string) interface{}
}
