package api

import (
	"bactic/internal/data"
	"bactic/internal/database"
	"context"
	"database/sql"
	"errors"
	"net/http"
)

// DefaultAPIService is a service that implements the logic for the DefaultAPIServicer
// This service should implement the business logic for every endpoint for the DefaultAPI API.
// Include any external packages or services that will be required by this service.
type APIServicer struct {
	searchTrie *Trie
	db         *sql.DB
}

// NewDefaultAPIService creates a default api service
func NewAPIServicer(dbURI string) *APIServicer {
	db := database.NewBacticDB("postgres", dbURI)
	trie := NewTrie()
	trie.CaseInsensitive()
	trie.WithNorm()

	rows, err := db.Query("SELECT name, id FROM athlete")
	if err != nil {
		panic(err)
	}

	var (
		i     data.SearchItem
		items []data.SearchItem
	)

	for rows.Next() {
		rows.Scan(&i.Name, &i.Id)
		i.ItemType = 0
		items = append(items, i)
	}

	trie.Insert(items...)
	return &APIServicer{
		searchTrie: trie,
		db:         db,
	}
}

// SearchAthleteGet -
func (s *APIServicer) SearchAthleteGet(ctx context.Context, name string) (ImplResponse, error) {
	return Response(http.StatusOK, s.searchTrie.Search(name, 10)), nil
}

// StatsAthleteIdGet -
func (s *APIServicer) StatsAthleteIdGet(ctx context.Context, id uint32) (ImplResponse, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Response(http.StatusInternalServerError, nil), errors.New("SQL database transaction could not begin")
	}

	summary, found := database.GetAthleteSummary(tx, id)
	if !found {
		return Response(http.StatusNotFound, nil), err
	}

	return Response(http.StatusOK, summary), nil
}

// StatsHistGet -
func (s *APIServicer) StatsHistGet(ctx context.Context, events []data.EventType, buckets float32) (ImplResponse, error) {
	// TODO - update StatsHistGet with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, []float32{}) or use other options such as http.Ok ...
	// return Response(200, []float32{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("StatsHistGet method not implemented")
}

// StatsTeamIdGet -
func (s *APIServicer) StatsTeamIdGet(ctx context.Context, id int64) (ImplResponse, error) {
	// TODO - update StatsTeamIdGet with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, TeamSummary{}) or use other options such as http.Ok ...
	// return Response(200, TeamSummary{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("StatsTeamIdGet method not implemented")
}

// StatsTimeseriesGet -
func (s *APIServicer) StatsTimeseriesGet(ctx context.Context, start string, end string, event int64) (ImplResponse, error) {
	// TODO - update StatsTimeseriesGet with the required logic for this service method.
	// Add api_default_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, StatsTimeseriesGet200Response{}) or use other options such as http.Ok ...
	// return Response(200, StatsTimeseriesGet200Response{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("StatsTimeseriesGet method not implemented")
}
