package api

import (
	"net/http"
)

func RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	
	// API endpoint
	mux.HandleFunc("POST /api/runs", HandleRun)

	// Serve frontend static files from the "web" folder
	fileServer := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fileServer)

	return mux
}
