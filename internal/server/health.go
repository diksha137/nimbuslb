package server

import (
	"encoding/json"
	"net/http"

	"github.com/diksha137/nimbuslb/internal/backend"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthHandler(backends []*backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, b := range backends {
			if b.IsHealthy() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				_ = json.NewEncoder(w).Encode(
					HealthResponse{
						Status: "healthy",
					},
				)

				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)

		_ = json.NewEncoder(w).Encode(
			HealthResponse{
				Status: "unhealthy",
			},
		)
	}
}
