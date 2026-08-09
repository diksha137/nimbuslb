package main

import (
	"log"
	"net/http"
	"time"

	"github.com/diksha137/nimbuslb/internal/backend"
	"github.com/diksha137/nimbuslb/internal/balancer"
	"github.com/diksha137/nimbuslb/internal/health"
)

func main() {

	backendA, err := backend.New(
		"Backend A",
		"http://localhost:9001",
	)
	if err != nil {
		log.Fatal(err)
	}

	backendB, err := backend.New(
		"Backend B",
		"http://localhost:9002",
	)
	if err != nil {
		log.Fatal(err)
	}

	backends := []*backend.Backend{
		backendA,
		backendB,
	}

	lb := balancer.New(backends)

	healthChecker := health.NewChecker(
		backends,
		5*time.Second,
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

	log.Println("NimbusLB listening on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
