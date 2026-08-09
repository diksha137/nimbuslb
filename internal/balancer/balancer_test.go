package balancer

import (
	"testing"

	"github.com/diksha137/nimbuslb/internal/backend"
)

func TestRoundRobin(t *testing.T) {
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

	backends := []*backend.Backend{
		backendA,
		backendB,
	}

	b := New(backends)

	expected := []string{
		"Backend A",
		"Backend B",
		"Backend A",
		"Backend B",
	}

	for i, expectedName := range expected {
		selected := b.NextBackend()

		if selected == nil {
			t.Fatalf("request %d returned nil backend", i)
		}

		if selected.Name != expectedName {
			t.Fatalf(
				"request %d: expected %s, got %s",
				i,
				expectedName,
				selected.Name,
			)
		}
	}
}

func TestSkipsUnhealthyBackend(t *testing.T) {
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

	backendB.SetHealthy(false)

	backends := []*backend.Backend{
		backendA,
		backendB,
	}

	b := New(backends)

	for i := 0; i < 10; i++ {
		selected := b.NextBackend()

		if selected == nil {
			t.Fatalf("request %d returned nil backend", i)
		}

		if selected.Name != "Backend A" {
			t.Fatalf(
				"request %d: expected Backend A, got %s",
				i,
				selected.Name,
			)
		}
	}
}

func TestReturnsNilWhenAllBackendsUnhealthy(t *testing.T) {
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

	backends := []*backend.Backend{
		backendA,
		backendB,
	}

	b := New(backends)

	selected := b.NextBackend()

	if selected != nil {
		t.Fatalf(
			"expected nil when all backends are unhealthy, got %s",
			selected.Name,
		)
	}
}
