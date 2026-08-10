package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diksha137/nimbuslb/internal/backend"
)

func TestHealthHandlerHealthy(t *testing.T) {
	backendA, err := backend.New(
		"Backend A",
		"http://localhost:9001",
	)
	if err != nil {
		t.Fatal(err)
	}

	backendB, err := backend.New(
		"Backend B",
		"http://localhost:9002",
	)
	if err != nil {
		t.Fatal(err)
	}

	backendA.SetHealthy(true)
	backendB.SetHealthy(true)

	handler := HealthHandler(
		[]*backend.Backend{
			backendA,
			backendB,
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var response HealthResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Fatalf(
			"expected status healthy, got %q",
			response.Status,
		)
	}

	if response.Backends["Backend A"] != "healthy" {
		t.Fatalf(
			"expected Backend A to be healthy, got %q",
			response.Backends["Backend A"],
		)
	}

	if response.Backends["Backend B"] != "healthy" {
		t.Fatalf(
			"expected Backend B to be healthy, got %q",
			response.Backends["Backend B"],
		)
	}
}

func TestHealthHandlerDegraded(t *testing.T) {
	backendA, err := backend.New(
		"Backend A",
		"http://localhost:9001",
	)
	if err != nil {
		t.Fatal(err)
	}

	backendB, err := backend.New(
		"Backend B",
		"http://localhost:9002",
	)
	if err != nil {
		t.Fatal(err)
	}

	backendA.SetHealthy(true)
	backendB.SetHealthy(false)

	handler := HealthHandler(
		[]*backend.Backend{
			backendA,
			backendB,
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d for degraded service, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var response HealthResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "degraded" {
		t.Fatalf(
			"expected status degraded, got %q",
			response.Status,
		)
	}

	if response.Backends["Backend A"] != "healthy" {
		t.Fatalf(
			"expected Backend A to be healthy, got %q",
			response.Backends["Backend A"],
		)
	}

	if response.Backends["Backend B"] != "unhealthy" {
		t.Fatalf(
			"expected Backend B to be unhealthy, got %q",
			response.Backends["Backend B"],
		)
	}
}

func TestHealthHandlerUnhealthy(t *testing.T) {
	handler := HealthHandler(
		[]*backend.Backend{},
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}

	var response HealthResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "unhealthy" {
		t.Fatalf(
			"expected status unhealthy, got %q",
			response.Status,
		)
	}

	if len(response.Backends) != 0 {
		t.Fatalf(
			"expected no backend statuses, got %d",
			len(response.Backends),
		)
	}
}
