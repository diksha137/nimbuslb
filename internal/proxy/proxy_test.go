package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReverseProxyReturnsBadGatewayOnBackendFailure(t *testing.T) {
	backend := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(
				w,
				"backend failure",
				http.StatusInternalServerError,
			)
		}),
	)

	backendURL := backend.URL
	backend.Close()

	proxy, err := NewReverseProxy(backendURL)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	recorder := httptest.NewRecorder()

	proxy.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadGateway,
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"Bad Gateway",
	) {
		t.Fatalf(
			"expected Bad Gateway response, got %q",
			recorder.Body.String(),
		)
	}
}
