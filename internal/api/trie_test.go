package api_test

import (
	"bactic/internal/api"
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
	entries := make([]string, nEntries)
	for i := 0; i < nEntries; i++ {
		entries = append(entries, uuid.New().String())
	}

	tr.Insert(entries...)
}

func TestTrieSearchMany(t *testing.T) {
	tr := api.NewTrie()
	nEntries := 100
	entries := make([]string, 0, nEntries)
	for i := 0; i < nEntries; i++ {
		entries = append(entries, uuid.New().String())
	}

	tr.Insert(entries...)

	values := tr.Search(entries[0][:1], 100)
	fmt.Println(values)
	valuesNames := make([]string, 0)

	for _, v := range values {
		valuesNames = append(valuesNames, v.Name)
	}

	if !slices.Contains(valuesNames, entries[0]) {
		t.Fatal("The returned values do not contain the searched string")
	}
}

func TestTrieSearchSingle(t *testing.T) {
	tr := api.NewTrie()
	nEntries := 100
	entries := make([]string, 0, nEntries)
	for i := 0; i < nEntries; i++ {
		entries = append(entries, uuid.New().String())
	}

	tr.Insert(entries...)

	values := tr.Search(entries[0][:5], 100)
	fmt.Println(values)
	valuesNames := make([]string, 0)

	for _, v := range values {
		valuesNames = append(valuesNames, v.Name)
	}

	if !slices.Contains(valuesNames, entries[0]) {
		t.Fatal("The returned values do not contain the searched string")
	}
}

func TestCreateMemoryProfile(t *testing.T) {
	tr := api.NewTrie()
	nEntries := 1_000_000
	entries := make([]string, 0, nEntries)
	for i := 0; i < nEntries; i++ {
		entries = append(entries, uuid.New().String())
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

	values := tr.Search(entries[0][:5], 10)
	pprof.StopCPUProfile()
	pprof.WriteHeapProfile(mf)
	valuesNames := make([]string, 0)

	for _, v := range values {
		valuesNames = append(valuesNames, v.Name)
	}

	if !slices.Contains(valuesNames, entries[0]) {
		t.Fatal("The returned values do not contain the searched string")
	}
}
