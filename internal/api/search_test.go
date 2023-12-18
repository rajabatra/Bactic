package api_test

import (
	"bactic/internal"
	"bactic/internal/api"
	"testing"
)

func TestBuildAthleteTrie(t *testing.T) {
	athletes := []internal.Athlete{
		{ID: 1, Name: "test1"},
		{ID: 2, Name: "abcedf"},
		{ID: 3, Name: "test2"},
		{ID: 4, Name: "hijlkmn"},
	}

	api.BuildAthleteTrie(athletes)
}

func TestSearchAthletes(t *testing.T) {
	athletes := []internal.Athlete{
		{ID: 1, Name: "test1"},
		{ID: 2, Name: "abcedf"},
		{ID: 3, Name: "test2"},
		{ID: 4, Name: "hijlkmn"},
	}

	root := api.BuildAthleteTrie(athletes)

	api.GetResults("test", 1, root)

}
