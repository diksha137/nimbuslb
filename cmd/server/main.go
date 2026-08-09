package main

import (
	"log"
	"net/http"

	"github.com/diksha137/nimbuslb/internal/backend"
	"github.com/diksha137/nimbuslb/internal/balancer"
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

	lb := balancer.New([]*backend.Backend{
		backendA,
		backendB,
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		selected := lb.NextBackend()

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
