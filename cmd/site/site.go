package main

import (
	"log"
	"net/http"
	"os"
)

var (
	logger *log.Logger
	srv    http.Server
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	logger.Println("Got one API request")
}

func main() {
	logger = log.Default()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", apiHandler)
	fs := http.FileServer(http.Dir("./static"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8080"
	}

	logger.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
