package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	server := &http.Server{
		Addr:    address,
		Handler: mux,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf(
			"NimbusLB listening on %s",
			address,
		)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf(
			"Graceful shutdown failed: %v",
			err,
		)
	}

	log.Println("NimbusLB shutdown complete")
}
