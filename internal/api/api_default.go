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
	"bactic/internal"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// DefaultAPIController binds http requests to an api service and writes the service results to the http response
type DefaultAPIController struct {
	service      DefaultAPIServicer
	errorHandler internal.ErrorHandler
}

// DefaultAPIOption for how the controller is set up.
type DefaultAPIOption func(*DefaultAPIController)

// WithDefaultAPIinternal.ErrorHandler inject internal.ErrorHandler into controller
func WithDefaultAPIErrorHandler(h internal.ErrorHandler) DefaultAPIOption {
	return func(c *DefaultAPIController) {
		c.errorHandler = h
	}
}

// NewDefaultAPIController creates a default api controller
func NewDefaultAPIController(s DefaultAPIServicer, opts ...DefaultAPIOption) internal.Router {
	controller := &DefaultAPIController{
		service:      s,
		errorHandler: internal.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the DefaultAPIController
func (c *DefaultAPIController) Routes() internal.Routes {
	return internal.Routes{
		"SearchAthleteGet": internal.Route{
			HandlerFunc: c.SearchAthleteGet,
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/search/athlete",
		},
		"StatsAthleteIdGet": internal.Route{
			HandlerFunc: c.StatsAthleteIdGet,
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/stats/athlete/{id}",
		},
		"StatsHistGet": internal.Route{
			HandlerFunc: c.StatsHistGet,
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/stats/hist",
		},
		"StatsTeamIdGet": internal.Route{
			HandlerFunc: c.StatsTeamIdGet,
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/stats/team/{id}",
		},
		"StatsTimeseriesGet": internal.Route{
			HandlerFunc: c.StatsTimeseriesGet,
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/stats/timeseries",
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
	internal.EncodeJSONResponse(result.Body, &result.Code, w)
}

// StatsAthleteIdGet -
func (c *DefaultAPIController) StatsAthleteIdGet(w http.ResponseWriter, r *http.Request) {
	idParam, err := internal.ParseNumericParameter[int64](
		chi.URLParam(r, "id"),
		internal.WithRequire[int64](internal.ParseInt64),
	)
	if err != nil {
		c.errorHandler(w, r, &internal.ParsingError{Err: err}, nil)
		return
	}
	result, err := c.service.StatsAthleteIdGet(r.Context(), idParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	internal.EncodeJSONResponse(result.Body, &result.Code, w)
}

// StatsHistGet -
func (c *DefaultAPIController) StatsHistGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var eventsParam []internal.EventType
	if query.Has("events") {
		paramSplits := strings.Split(query.Get("events"), ",")
		eventsParam = make([]internal.EventType, 0, len(paramSplits))
		for _, param := range paramSplits {
			paramEnum, err := internal.NewEventTypeFromValue(param)
			if err != nil {
				c.errorHandler(w, r, &internal.ParsingError{Err: err}, nil)
				return
			}
			eventsParam = append(eventsParam, paramEnum)
		}
	}
	var bucketsParam float32
	if query.Has("buckets") {
		param, err := internal.ParseNumericParameter[float32](
			query.Get("buckets"),
			internal.WithParse[float32](internal.ParseFloat32),
		)
		if err != nil {
			c.errorHandler(w, r, &internal.ParsingError{Err: err}, nil)
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
	internal.EncodeJSONResponse(result.Body, &result.Code, w)
}

// StatsTeamIdGet -
func (c *DefaultAPIController) StatsTeamIdGet(w http.ResponseWriter, r *http.Request) {
	idParam, err := internal.ParseNumericParameter[int64](
		chi.URLParam(r, "id"),
		internal.WithRequire[int64](internal.ParseInt64),
	)
	if err != nil {
		c.errorHandler(w, r, &internal.ParsingError{Err: err}, nil)
		return
	}
	result, err := c.service.StatsTeamIdGet(r.Context(), idParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	internal.EncodeJSONResponse(result.Body, &result.Code, w)
}

// StatsTimeseriesGet -
func (c *DefaultAPIController) StatsTimeseriesGet(w http.ResponseWriter, r *http.Request) {
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
		param, err := internal.ParseNumericParameter[int64](
			query.Get("event"),
			internal.WithParse[int64](internal.ParseInt64),
		)
		if err != nil {
			c.errorHandler(w, r, &internal.ParsingError{Err: err}, nil)
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
	internal.EncodeJSONResponse(result.Body, &result.Code, w)
}
