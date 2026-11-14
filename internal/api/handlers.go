package api

import (
	"encoding/json"
	"lua-runner/internal/execution"
	"net/http"
	"time"
)


type RunRequest struct {
	Code string `json:"code"`
}

func HandleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit body size to 64KB
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		http.Error(w, "Invalid request body or empty code", http.StatusBadRequest)
		return
	}

	result, err := execution.RunLua(req.Code, 2*time.Second)
	if err != nil {
		http.Error(w, "Execution error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Connect-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
