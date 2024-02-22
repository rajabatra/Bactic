/* Provides the actual routing implementation for this api */
package api

import (
	"bactic/internal"
	"bactic/internal/data"
	"net/http"
	"strings"
)

// DefaultAPIController binds http requests to an api service and writes the service results to the http response
type APIRouter struct {
	service      *APIServicer
	errorHandler ErrorHandler
}

// NewDefaultAPIController creates a default api controller
func NewAPIRouter(s *APIServicer) *APIRouter {
	router := &APIRouter{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}
	return router
}

// Routes returns all the api routes for the APIRouter, making this a Router
func (c *APIRouter) Routes() Routes {
	return Routes{
		"SearchAthleteGet": Route{
			HandlerFunc: c.SearchAthleteGet,
			Method:      "GET",
			Pattern:     "/search/athlete",
		},
		"StatsAthleteIdGet": Route{
			HandlerFunc: c.StatsAthleteIdGet,
			Method:      "GET",
			Pattern:     "/stats/athlete/{id}",
		},
		"StatsHistGet": Route{
			HandlerFunc: c.StatsHistGet,
			Method:      "GET",
			Pattern:     "/stats/hist",
		},
		"StatsTeamIdGet": Route{
			HandlerFunc: c.StatsTeamIdGet,
			Method:      "GET",
			Pattern:     "/stats/team/{id}",
		},
		"StatsTimeseriesGet": Route{
			HandlerFunc: c.StatsTimeseriesGet,
			Method:      "GET",
			Pattern:     "/stats/timeseries",
		},
	}
}

// SearchAthleteGet -
func (c *APIRouter) SearchAthleteGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var nameParam string
	if query.Has("name") {
		param := query.Get("name")

		nameParam = param
	} else {
		c.errorHandler(w, r, &internal.RequiredError{Field: "name"}, nil)
		return
	}
	result, err := c.service.SearchAthleteGet(r.Context(), nameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// StatsAthleteIdGet -
func (c *APIRouter) StatsAthleteIdGet(w http.ResponseWriter, r *http.Request) {
	idParam, err := ParseNumericParameter[uint32](
		r.PathValue("id"),
		WithRequire[uint32](ParseUint32),
	)
	if err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	result, err := c.service.StatsAthleteIdGet(r.Context(), idParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// StatsHistGet -
func (c *APIRouter) StatsHistGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var eventsParam []data.EventType
	if query.Has("events") {
		paramSplits := strings.Split(query.Get("events"), ",")
		eventsParam = make([]data.EventType, 0, len(paramSplits))
		for _, param := range paramSplits {
			paramEnum, err := data.NewEventTypeFromValue(param)
			if err != nil {
				c.errorHandler(w, r, &ParsingError{Err: err}, nil)
				return
			}
			eventsParam = append(eventsParam, paramEnum)
		}
	}
	var bucketsParam float32
	if query.Has("buckets") {
		param, err := ParseNumericParameter[float32](
			query.Get("buckets"),
			WithParse[float32](ParseFloat32),
		)
		if err != nil {
			c.errorHandler(w, r, &ParsingError{Err: err}, nil)
			return
		}

		bucketsParam = param
	} else {
		var param float32 = 10
		bucketsParam = param
	}
	result, err := c.service.StatsHistGet(r.Context(), eventsParam, bucketsParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// StatsTeamIdGet -
func (c *APIRouter) StatsTeamIdGet(w http.ResponseWriter, r *http.Request) {
	idParam, err := ParseNumericParameter[int64](
		r.PathValue("id"),
		WithRequire[int64](ParseInt64),
	)
	if err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	result, err := c.service.StatsTeamIdGet(r.Context(), idParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// StatsTimeseriesGet -
func (c *APIRouter) StatsTimeseriesGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var startParam string
	if query.Has("start") {
		param := query.Get("start")

		startParam = param
	} else {
		c.errorHandler(w, r, &internal.RequiredError{Field: "start"}, nil)
		return
	}
	var endParam string
	if query.Has("end") {
		param := query.Get("end")

		endParam = param
	} else {
		c.errorHandler(w, r, &internal.RequiredError{Field: "end"}, nil)
		return
	}
	var eventParam int64
	if query.Has("event") {
		param, err := ParseNumericParameter[int64](
			query.Get("event"),
			WithParse[int64](ParseInt64),
		)
		if err != nil {
			c.errorHandler(w, r, &ParsingError{Err: err}, nil)
			return
		}

		eventParam = param
	}
	result, err := c.service.StatsTimeseriesGet(r.Context(), startParam, endParam, eventParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}
