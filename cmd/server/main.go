package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	// Public demo dashboard.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		serveDashboard(w)
	})

	// Actual load-balanced request endpoint.
	mux.HandleFunc("/demo/request", func(w http.ResponseWriter, r *http.Request) {
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

	port := os.Getenv("PORT")

	if port == "" {
		port = fmt.Sprintf("%d", cfg.Server.Port)
	}

	address := ":" + port

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

func serveDashboard(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	const dashboard = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">

	<title>NimbusLB — Live Load Balancer Demo</title>

	<style>
		* {
			box-sizing: border-box;
		}

		body {
			margin: 0;
			font-family:
				Inter,
				system-ui,
				-apple-system,
				BlinkMacSystemFont,
				"Segoe UI",
				sans-serif;

			background: #0f172a;
			color: #e2e8f0;
		}

		.container {
			max-width: 1000px;
			margin: 0 auto;
			padding: 48px 20px;
		}

		.hero {
			margin-bottom: 32px;
		}

		.badge {
			display: inline-block;
			padding: 6px 12px;
			border-radius: 999px;
			background: #1e293b;
			color: #93c5fd;
			font-size: 13px;
			font-weight: 600;
			margin-bottom: 16px;
		}

		h1 {
			font-size: clamp(38px, 7vw, 64px);
			margin: 0 0 12px;
			letter-spacing: -2px;
		}

		.subtitle {
			font-size: 20px;
			color: #94a3b8;
			max-width: 760px;
			line-height: 1.6;
		}

		.grid {
			display: grid;
			grid-template-columns: repeat(2, 1fr);
			gap: 18px;
			margin-bottom: 18px;
		}

		.card {
			background: #1e293b;
			border: 1px solid #334155;
			border-radius: 16px;
			padding: 24px;
		}

		.card h2 {
			margin-top: 0;
			font-size: 18px;
		}

		.explanation {
			line-height: 1.7;
			color: #cbd5e1;
		}

		.health {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 14px 0;
			border-bottom: 1px solid #334155;
		}

		.health:last-child {
			border-bottom: 0;
		}

		.status {
			display: flex;
			align-items: center;
			gap: 8px;
			font-weight: 600;
		}

		.dot {
			width: 10px;
			height: 10px;
			border-radius: 50%;
			display: inline-block;
			background: #94a3b8;
		}

		.dot.healthy {
			background: #22c55e;
		}

		.dot.unhealthy {
			background: #ef4444;
		}

		.button {
			display: inline-block;
			border: 0;
			border-radius: 10px;
			padding: 13px 20px;
			background: #2563eb;
			color: white;
			font-weight: 700;
			font-size: 15px;
			cursor: pointer;
		}

		.button:hover {
			background: #1d4ed8;
		}

		.button:disabled {
			opacity: 0.6;
			cursor: wait;
		}

		.result {
			margin-top: 16px;
			padding: 14px;
			border-radius: 10px;
			background: #0f172a;
			color: #93c5fd;
			min-height: 48px;
		}

		.metrics {
			display: grid;
			grid-template-columns: repeat(4, 1fr);
			gap: 12px;
		}

		.metric {
			background: #0f172a;
			border-radius: 12px;
			padding: 16px;
		}

		.metric-value {
			font-size: 28px;
			font-weight: 700;
			margin-bottom: 4px;
		}

		.metric-label {
			color: #94a3b8;
			font-size: 13px;
		}

		.links {
			display: flex;
			flex-wrap: wrap;
			gap: 12px;
			margin-top: 18px;
		}

		a {
			color: #93c5fd;
			text-decoration: none;
		}

		a:hover {
			text-decoration: underline;
		}

		code {
			background: #0f172a;
			padding: 3px 7px;
			border-radius: 6px;
		}

		.footer {
			margin-top: 32px;
			color: #64748b;
			font-size: 14px;
			text-align: center;
		}

		@media (max-width: 700px) {
			.grid {
				grid-template-columns: 1fr;
			}

			.metrics {
				grid-template-columns: repeat(2, 1fr);
			}
		}
	</style>
</head>

<body>

<div class="container">

	<section class="hero">
		<div class="badge">LIVE LOAD BALANCER DEMO</div>

		<h1>NimbusLB</h1>

		<div class="subtitle">
			A production-oriented HTTP load balancer written in Go.
			NimbusLB distributes traffic across multiple backend servers
			using round-robin scheduling, continuously monitors backend health,
			and exposes operational metrics.
		</div>
	</section>

	<div class="grid">

		<section class="card">
			<h2>How the demo works</h2>

			<p class="explanation">
				Your request first reaches <strong>NimbusLB</strong>.
				The load balancer then selects a healthy backend and
				forwards the request to it.
			</p>

			<p class="explanation">
				With two healthy backends, requests are distributed using
				<strong>round-robin scheduling</strong>.
			</p>

			<p class="explanation">
				Try the button several times and watch the response alternate
				between Backend A and Backend B.
			</p>

			<button
				class="button"
				id="requestButton"
				onclick="sendRequest()"
			>
				Send Test Request
			</button>

			<div class="result" id="result">
				No test request sent yet.
			</div>
		</section>

		<section class="card">
			<h2>Backend Health</h2>

			<div class="health">
				<span>Backend A</span>

				<span class="status">
					<span class="dot" id="backendADot"></span>
					<span id="backendAStatus">Checking...</span>
				</span>
			</div>

			<div class="health">
				<span>Backend B</span>

				<span class="status">
					<span class="dot" id="backendBDot"></span>
					<span id="backendBStatus">Checking...</span>
				</span>
			</div>

			<p class="explanation">
				NimbusLB periodically checks backend availability and
				removes unhealthy instances from the rotation.
			</p>

			<a href="/health" target="_blank">
				View raw health response →
			</a>
		</section>

	</div>

	<section class="card">
		<h2>Live Metrics</h2>

		<div class="metrics">

			<div class="metric">
				<div class="metric-value" id="requests">—</div>
				<div class="metric-label">Total Requests</div>
			</div>

			<div class="metric">
				<div class="metric-value" id="success">—</div>
				<div class="metric-label">Successful</div>
			</div>

			<div class="metric">
				<div class="metric-value" id="failed">—</div>
				<div class="metric-label">Failed</div>
			</div>

			<div class="metric">
				<div class="metric-value" id="backendA">—</div>
				<div class="metric-label">Backend A</div>
			</div>

		</div>

		<p class="explanation">
			Backend B requests:
			<strong id="backendB">—</strong>
		</p>

		<a href="/metrics" target="_blank">
			View raw metrics →
		</a>
	</section>

	<section class="card" style="margin-top: 18px;">

		<h2>What this project demonstrates</h2>

		<p class="explanation">
			NimbusLB is a small but complete networking system built in Go.
			It demonstrates:
		</p>

		<p class="explanation">
			<strong>HTTP reverse proxying</strong> ·
			<strong>round-robin load balancing</strong> ·
			<strong>health checking</strong> ·
			<strong>failure handling</strong> ·
			<strong>concurrency</strong> ·
			<strong>request tracing</strong> ·
			<strong>metrics</strong> ·
			<strong>graceful shutdown</strong> ·
			<strong>Docker deployment</strong>
		</p>

		<div class="links">
			<a href="/health" target="_blank">Health Endpoint</a>
			<a href="/metrics" target="_blank">Metrics Endpoint</a>

			<a
				href="https://github.com/diksha137/nimbuslb"
				target="_blank"
				rel="noopener noreferrer"
			>
				GitHub Source Code
			</a>
		</div>

	</section>

	<div class="footer">
		NimbusLB · Go · HTTP · Reverse Proxy · Load Balancing
	</div>

</div>

<script>
	function parseMetric(text, name, labels) {
	const lines = text.split("\n");

	for (const line of lines) {
		if (!line.startsWith(name)) {
			continue;
		}

		if (labels && !line.includes(labels)) {
			continue;
		}

		const match = line.match(/\}\s+([0-9.]+)$/);

		if (match) {
			return match[1];
		}

		const parts = line.trim().split(/\s+/);

		if (parts.length >= 2) {
			return parts[parts.length - 1];
		}
	}

	return "0";
}

	async function refreshMetrics() {
		try {
			const response = await fetch("/metrics");

			if (!response.ok) {
				throw new Error("Metrics request failed");
			}

			const text = await response.text();

			document.getElementById("requests").textContent =
				parseMetric(
					text,
					"nimbuslb_requests_total"
				);

			document.getElementById("success").textContent =
				parseMetric(
					text,
					"nimbuslb_requests_success_total"
				);

			document.getElementById("failed").textContent =
				parseMetric(
					text,
					"nimbuslb_requests_failed_total"
				);

			document.getElementById("backendA").textContent =
				parseMetric(
					text,
					"nimbuslb_backend_requests_total",
					'backend="Backend A"'
				);

			document.getElementById("backendB").textContent =
				parseMetric(
					text,
					"nimbuslb_backend_requests_total",
					'backend="Backend B"'
				);

		} catch (error) {
			console.error("Failed to load metrics:", error);
		}
	}

async function refreshHealth() {
    try {
        const response = await fetch("/health");

        if (!response.ok) {
            throw new Error("Health request failed");
        }

        const data = await response.json();

        const backendAHealthy =
            data.backends &&
            data.backends["Backend A"] === "healthy";

        const backendBHealthy =
            data.backends &&
            data.backends["Backend B"] === "healthy";

        updateHealth(
            "backendADot",
            "backendAStatus",
            backendAHealthy
        );

        updateHealth(
            "backendBDot",
            "backendBStatus",
            backendBHealthy
        );

    } catch (error) {
        console.error("Failed to load health:", error);

        updateHealth(
            "backendADot",
            "backendAStatus",
            false,
            "Unavailable"
        );

        updateHealth(
            "backendBDot",
            "backendBStatus",
            false,
            "Unavailable"
        );
    }
}

	function updateHealth(dotID, statusID, healthy, label) {
		const dot = document.getElementById(dotID);
		const status = document.getElementById(statusID);

		dot.classList.remove("healthy");
		dot.classList.remove("unhealthy");

		if (healthy) {
			dot.classList.add("healthy");
			status.textContent = "Healthy";
		} else {
			dot.classList.add("unhealthy");
			status.textContent = label || "Unhealthy";
		}
	}

	async function sendRequest() {
		const result = document.getElementById("result");
		const button = document.getElementById("requestButton");

		button.disabled = true;
		result.textContent = "Sending request through NimbusLB...";

		try {
			const response = await fetch("/demo/request");

			const text = await response.text();

			if (!response.ok) {
				throw new Error(text.trim() || "Request failed");
			}

			result.textContent =
				"Load-balanced response: " + text.trim();

			await refreshMetrics();

		} catch (error) {
			result.textContent =
				"Request failed: " + error.message;
		} finally {
			button.disabled = false;
		}
	}

	refreshMetrics();
	refreshHealth();

	setInterval(refreshMetrics, 3000);
	setInterval(refreshHealth, 5000);
</script>

</body>
</html>`

	if _, err := w.Write(
		[]byte(strings.TrimSpace(dashboard)),
	); err != nil {
		log.Printf("failed to write dashboard: %v", err)
	}
}
