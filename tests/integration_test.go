package tests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diksha137/nimbuslb/internal/backend"
	"github.com/diksha137/nimbuslb/internal/balancer"
)

func createTestBackend(t *testing.T, name string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello from "+name)
		}),
	)
}

func TestLoadBalancerHTTPRouting(t *testing.T) {
	serverA := createTestBackend(t, "Backend A")
	defer serverA.Close()

	serverB := createTestBackend(t, "Backend B")
	defer serverB.Close()

	backendA, err := backend.New(
		"Backend A",
		serverA.URL,
	)
	if err != nil {
		t.Fatal(err)
	}

	backendB, err := backend.New(
		"Backend B",
		serverB.URL,
	)
	if err != nil {
		t.Fatal(err)
	}

	lb := balancer.New(
		[]*backend.Backend{
			backendA,
			backendB,
		},
	)

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			selected := lb.NextBackend()

			if selected == nil {
				http.Error(
					w,
					"No healthy backends available",
					http.StatusServiceUnavailable,
				)
				return
			}

			selected.Proxy.ServeHTTP(w, r)
		},
	)

	lbServer := httptest.NewServer(handler)
	defer lbServer.Close()

	expected := []string{
		"Hello from Backend A",
		"Hello from Backend B",
		"Hello from Backend A",
		"Hello from Backend B",
	}

	for i, expectedResponse := range expected {
		resp, err := http.Get(lbServer.URL)
		if err != nil {
			t.Fatalf(
				"request %d failed: %v",
				i,
				err,
			)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			t.Fatalf(
				"request %d failed reading response: %v",
				i,
				err,
			)
		}

		actual := string(body)

		if actual != expectedResponse+"\n" {
			t.Fatalf(
				"request %d: expected %q, got %q",
				i,
				expectedResponse+"\n",
				actual,
			)
		}
	}
}

func TestLoadBalancerFailover(t *testing.T) {
	serverA := createTestBackend(t, "Backend A")
	defer serverA.Close()

	serverB := createTestBackend(t, "Backend B")
	defer serverB.Close()

	backendA, err := backend.New(
		"Backend A",
		serverA.URL,
	)
	if err != nil {
		t.Fatal(err)
	}

	backendB, err := backend.New(
		"Backend B",
		serverB.URL,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate Backend B becoming unhealthy.
	backendB.SetHealthy(false)

	lb := balancer.New(
		[]*backend.Backend{
			backendA,
			backendB,
		},
	)

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			selected := lb.NextBackend()

			if selected == nil {
				http.Error(
					w,
					"No healthy backends available",
					http.StatusServiceUnavailable,
				)
				return
			}

			selected.Proxy.ServeHTTP(w, r)
		},
	)

	lbServer := httptest.NewServer(handler)
	defer lbServer.Close()

	for i := 0; i < 10; i++ {
		resp, err := http.Get(lbServer.URL)
		if err != nil {
			t.Fatalf(
				"request %d failed: %v",
				i,
				err,
			)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			t.Fatalf(
				"request %d failed reading response: %v",
				i,
				err,
			)
		}

		expected := "Hello from Backend A\n"
		actual := string(body)

		if actual != expected {
			t.Fatalf(
				"request %d: expected %q, got %q",
				i,
				expected,
				actual,
			)
		}
	}
}

func TestLoadBalancerAllBackendsUnavailable(t *testing.T) {
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

	lb := balancer.New(
		[]*backend.Backend{
			backendA,
			backendB,
		},
	)

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			selected := lb.NextBackend()

			if selected == nil {
				http.Error(
					w,
					"No healthy backends available",
					http.StatusServiceUnavailable,
				)
				return
			}

			selected.Proxy.ServeHTTP(w, r)
		},
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			resp.StatusCode,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected := "No healthy backends available\n"

	if string(body) != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			string(body),
		)
	}
}
