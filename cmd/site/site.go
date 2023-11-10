package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	logger *log.Logger
)

// Root api handler function. All api requests are routed in here
func apiHandler(w http.ResponseWriter, r *http.Request) {
	logger.Println("Got one API request")
}

// Sets up a file server that serves ./web/index.html to the root page and then static assets to all other paths
func setupFileServer(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/" {
			http.ServeFile(w, r, "./web/index.html")
		} else {
			http.ServeFile(w, r, filepath.Join("./web", r.RequestURI))
		}
	})
}

func main() {
	logger = log.Default()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", apiHandler)
	setupFileServer(mux)

	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8080"
	}

	logger.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
