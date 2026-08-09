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
	"github.com/diksha137/nimbuslb/internal/metrics"
	"github.com/diksha137/nimbuslb/internal/middleware"
	serverhandlers "github.com/diksha137/nimbuslb/internal/server"
)

func main() {
	configPath := os.Getenv("NIMBUSLB_CONFIG")

	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
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
	metricsCollector := metrics.New()

	healthChecker := health.NewChecker(
		backends,
		time.Duration(cfg.Health.IntervalSeconds)*time.Second,
	)

	healthChecker.Start()

	mux := http.NewServeMux()
	mux.HandleFunc(
		"/metrics",
		metricsCollector.Handler,
	)

	mux.Handle(
		"/health",
		serverhandlers.HealthHandler(backends),
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		metricsCollector.IncRequests()

		selected := lb.NextBackend()

		if selected == nil {
			metricsCollector.IncFailed()

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
		metricsCollector.IncBackendRequest(selected.Name)

		selected.Proxy.ServeHTTP(w, r)
		metricsCollector.IncSuccess()
	})

	address := fmt.Sprintf(":%d", cfg.Server.Port)

	handler := middleware.RequestID(
		middleware.Logging(mux),
	)

	server := &http.Server{
		Addr:    address,
		Handler: handler,

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
