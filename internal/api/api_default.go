/*
 * Bactic
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.1.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// DefaultAPIController binds http requests to an api service and writes the service results to the http response
type DefaultAPIController struct {
	service      DefaultAPIServicer
	errorHandler ErrorHandler
}

// DefaultAPIOption for how the controller is set up.
type DefaultAPIOption func(*DefaultAPIController)

// WithDefaultAPIErrorHandler inject ErrorHandler into controller
func WithDefaultAPIErrorHandler(h ErrorHandler) DefaultAPIOption {
	return func(c *DefaultAPIController) {
		c.errorHandler = h
	}
}

// NewDefaultAPIController creates a default api controller
func NewDefaultAPIController(s DefaultAPIServicer, opts ...DefaultAPIOption) Router {
	controller := &DefaultAPIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the DefaultAPIController
func (c *DefaultAPIController) Routes() Routes {
	return Routes{
		"SearchAthleteGet": Route{
			c.SearchAthleteGet,
			strings.ToUpper("Get"),
			"/search/athlete",
		},
		"StatsAthleteIdGet": Route{
			c.StatsAthleteIdGet,
			strings.ToUpper("Get"),
			"/stats/athlete/{id}",
		},
		"StatsHistGet": Route{
			c.StatsHistGet,
			strings.ToUpper("Get"),
			"/stats/hist",
		},
		"StatsTeamIdGet": Route{
			c.StatsTeamIdGet,
			strings.ToUpper("Get"),
			"/stats/team/{id}",
		},
		"StatsTimeseriesGet": Route{
			c.StatsTimeseriesGet,
			strings.ToUpper("Get"),
			"/stats/timeseries",
		},
	}
}

// SearchAthleteGet -
func (c *DefaultAPIController) SearchAthleteGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var nameParam string
	if query.Has("name") {
		param := query.Get("name")

		nameParam = param
	} else {
		c.errorHandler(w, r, &RequiredError{Field: "name"}, nil)
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
func (c *DefaultAPIController) StatsAthleteIdGet(w http.ResponseWriter, r *http.Request) {
	idParam, err := parseNumericParameter[int64](
		chi.URLParam(r, "id"),
		WithRequire[int64](parseInt64),
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
func (c *DefaultAPIController) StatsHistGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var eventsParam []Event
	if query.Has("events") {
		paramSplits := strings.Split(query.Get("events"), ",")
		eventsParam = make([]Event, 0, len(paramSplits))
		for _, param := range paramSplits {
			paramEnum, err := NewEventFromValue(param)
			if err != nil {
				c.errorHandler(w, r, &ParsingError{Err: err}, nil)
				return
			}
			eventsParam = append(eventsParam, paramEnum)
		}
	}
	var bucketsParam float32
	if query.Has("buckets") {
		param, err := parseNumericParameter[float32](
			query.Get("buckets"),
			WithParse[float32](parseFloat32),
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
func (c *DefaultAPIController) StatsTeamIdGet(w http.ResponseWriter, r *http.Request) {
	idParam, err := parseNumericParameter[int64](
		chi.URLParam(r, "id"),
		WithRequire[int64](parseInt64),
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
func (c *DefaultAPIController) StatsTimeseriesGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var startParam string
	if query.Has("start") {
		param := query.Get("start")

		startParam = param
	} else {
		c.errorHandler(w, r, &RequiredError{Field: "start"}, nil)
		return
	}
	var endParam string
	if query.Has("end") {
		param := query.Get("end")

		endParam = param
	} else {
		c.errorHandler(w, r, &RequiredError{Field: "end"}, nil)
		return
	}
	var eventParam int64
	if query.Has("event") {
		param, err := parseNumericParameter[int64](
			query.Get("event"),
			WithParse[int64](parseInt64),
		)
		if err != nil {
			c.errorHandler(w, r, &ParsingError{Err: err}, nil)
			return
		}

		eventParam = param
	} else {
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
