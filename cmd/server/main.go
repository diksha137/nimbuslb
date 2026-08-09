package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/diksha137/nimbuslb/internal/backend"
	"github.com/diksha137/nimbuslb/internal/balancer"
	"github.com/diksha137/nimbuslb/internal/config"
	"github.com/diksha137/nimbuslb/internal/health"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var backends []*backend.Backend

	for _, backendConfig := range cfg.Backends {
		b, err := backend.New(
			backendConfig.Name,
			backendConfig.URL,
		)

		if err != nil {
			log.Fatalf(
				"failed to create backend %s: %v",
				backendConfig.Name,
				err,
			)
		}

		backends = append(backends, b)
	}

	lb := balancer.New(backends)

	healthChecker := health.NewChecker(
		backends,
		time.Duration(cfg.Health.IntervalSeconds)*time.Second,
	)

	healthChecker.Start()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		selected := lb.NextBackend()

		if selected == nil {
			http.Error(
				w,
				"No healthy backends available",
				http.StatusServiceUnavailable,
			)
			return
		}

		log.Printf(
			"Forwarding %s %s -> %s",
			r.Method,
			r.URL.Path,
			selected.Name,
		)

		selected.Proxy.ServeHTTP(w, r)
	})

	address := fmt.Sprintf(":%d", cfg.Server.Port)

	log.Printf(
		"NimbusLB listening on %s",
		address,
	)

	log.Fatal(http.ListenAndServe(address, nil))
}
