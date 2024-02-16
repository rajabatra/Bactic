package api_test

import (
	"bactic/internal/api"
	"bactic/internal/data"
	"fmt"
	"os"
	"runtime/pprof"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func TestNewTrie(t *testing.T) {
	tr := api.NewTrie()
	tr.CaseInsensitive()
	tr.WithoutNorm()
}

func TestTrieInsert(t *testing.T) {
	tr := api.NewTrie()

	nEntries := 100
	entries := make([]data.SearchItem, nEntries)
	for i := 0; i < nEntries; i++ {
		entries = append(entries, data.SearchItem{
			Name: uuid.NewString(),
		})
	}

	tr.Insert(entries...)
}

func TestTrieSearchMany(t *testing.T) {
	tr := api.NewTrie()
	nEntries := 100
	entries := make([]data.SearchItem, 0, nEntries)
	for i := 0; i < nEntries; i++ {
		entries = append(entries, data.SearchItem{
			Name: uuid.NewString(),
		})
	}

	tr.Insert(entries...)

	values := tr.Search(entries[0].Name[1:], 100)
	fmt.Println(values)
	valuesNames := make([]string, 0)

	for _, v := range values {
		valuesNames = append(valuesNames, v.Name)
	}

	if !slices.Contains(valuesNames, entries[0].Name) {
		t.Fatal("The returned values do not contain the searched string")
	}
}

func TestTrieSearchSingle(t *testing.T) {
	tr := api.NewTrie()
	nEntries := 100
	entries := make([]data.SearchItem, 0, nEntries)
	for i := 0; i < nEntries; i++ {
		entries = append(entries, data.SearchItem{
			Name: uuid.NewString(),
		})
	}

	tr.Insert(entries...)

	values := tr.Search(entries[0].Name[:5], 100)
	valuesNames := make([]string, 0)

	for _, v := range values {
		valuesNames = append(valuesNames, v.Name)
	}

	if !slices.Contains(valuesNames, entries[0].Name) {
		t.Fatal("The returned values do not contain the searched string")
	}
}

func TestCreateMemoryProfile(t *testing.T) {
	tr := api.NewTrie()
	nEntries := 1_000_000
	entries := make([]data.SearchItem, 0, nEntries)
	for i := 0; i < nEntries; i++ {
		entries = append(entries, data.SearchItem{
			Name: uuid.NewString(),
		})
	}

	mf, err := os.Create("./trie.mprof")
	if err != nil {
		panic(err)
	}
	cf, err := os.Create("./trie.cprof")
	if err != nil {
		panic(err)
	}
	defer mf.Close()
	defer cf.Close()
	// profile the expensive functions
	pprof.StartCPUProfile(cf)
	tr.Insert(entries...)

	values := tr.Search(entries[0].Name[:5], 10)
	pprof.StopCPUProfile()
	pprof.WriteHeapProfile(mf)
	valuesNames := make([]string, 0)

	for _, v := range values {
		valuesNames = append(valuesNames, v.Name)
	}

	if !slices.Contains(valuesNames, entries[0].Name) {
		t.Fatal("The returned values do not contain the searched string")
	}
}
