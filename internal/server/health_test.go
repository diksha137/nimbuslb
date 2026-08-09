package server

import (
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

	handler := HealthHandler(
		[]*backend.Backend{
			backendA,
			backendB,
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	expected := `{"status":"healthy"}
`

	if recorder.Body.String() != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			recorder.Body.String(),
		)
	}
}

func TestHealthHandlerUnhealthy(t *testing.T) {
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

	backendA.SetHealthy(false)
	backendB.SetHealthy(false)

	handler := HealthHandler(
		[]*backend.Backend{
			backendA,
			backendB,
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			recorder.Code,
		)
	}

	expected := `{"status":"unhealthy"}
`

	if recorder.Body.String() != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			recorder.Body.String(),
		)
	}
}
