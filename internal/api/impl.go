package api

// ImplResponse defines an implementation response with error code and the associated body
type ImplResponse struct {
	Body interface{}
	Code int
}
