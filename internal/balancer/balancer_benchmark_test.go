package balancer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diksha137/nimbuslb/internal/backend"
)

func BenchmarkNextBackend(b *testing.B) {
	backendA, err := backend.New(
		"Backend A",
		"http://localhost:9001",
	)
	if err != nil {
		b.Fatal(err)
	}

	backendB, err := backend.New(
		"Backend B",
		"http://localhost:9002",
	)
	if err != nil {
		b.Fatal(err)
	}

	backendA.SetHealthy(true)
	backendB.SetHealthy(true)

	lb := New([]*backend.Backend{
		backendA,
		backendB,
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lb.NextBackend()
	}
}

func BenchmarkHTTPRouting(b *testing.B) {
	backendServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	defer backendServer.Close()

	backendA, err := backend.New(
		"Backend A",
		backendServer.URL,
	)
	if err != nil {
		b.Fatal(err)
	}

	backendA.SetHealthy(true)

	lb := New([]*backend.Backend{
		backendA,
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		selected := lb.NextBackend()

		if selected == nil {
			b.Fatal("expected healthy backend")
		}

		req := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

		recorder := httptest.NewRecorder()

		selected.Proxy.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			b.Fatalf(
				"expected status 200, got %d",
				recorder.Code,
			)
		}
	}
}

func BenchmarkNextBackendParallel(b *testing.B) {
	backendA, err := backend.New(
		"Backend A",
		"http://localhost:9001",
	)
	if err != nil {
		b.Fatal(err)
	}

	backendB, err := backend.New(
		"Backend B",
		"http://localhost:9002",
	)
	if err != nil {
		b.Fatal(err)
	}

	backendA.SetHealthy(true)
	backendB.SetHealthy(true)

	lb := New([]*backend.Backend{
		backendA,
		backendB,
	})

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lb.NextBackend()
		}
	})
}
