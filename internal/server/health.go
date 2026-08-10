package server

import (
	"encoding/json"
	"net/http"

	"github.com/diksha137/nimbuslb/internal/backend"
)

type HealthResponse struct {
	Status   string            `json:"status"`
	Backends map[string]string `json:"backends"`
}

func HealthHandler(backends []*backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		allHealthy := true
		backendStatuses := make(map[string]string)

		for _, b := range backends {
			if b.IsHealthy() {
				backendStatuses[b.Name] = "healthy"
			} else {
				backendStatuses[b.Name] = "unhealthy"
				allHealthy = false
			}
		}

		status := "healthy"

		if !allHealthy {
			status = "degraded"
		}

		if len(backends) == 0 {
			status = "unhealthy"
		}

		if status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		_ = json.NewEncoder(w).Encode(
			HealthResponse{
				Status:   status,
				Backends: backendStatuses,
			},
		)
	}
}
