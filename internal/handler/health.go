package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the response body for health check
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
}

// healthHandler handles health check HTTP requests
type healthHandler struct {
	version string
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(version string) *healthHandler {
	return &healthHandler{
		version: version,
	}
}

// Health returns a handler function for health check
func (h *healthHandler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Create health response
		health := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Service:   "conduit-api",
			Version:   h.version,
		}

		// Set status code
		w.WriteHeader(http.StatusOK)

		// Encode and send response
		if err := json.NewEncoder(w).Encode(health); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
