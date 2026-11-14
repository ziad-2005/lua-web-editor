package main

import (
	"log"
	"lua-runner/internal/api"
	"net/http"
)


func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/runs", api.HandleRun)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed: %v", err)
	}
}
