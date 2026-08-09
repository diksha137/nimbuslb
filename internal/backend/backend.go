package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy
	Name  string

	mu      sync.RWMutex
	healthy bool
}

func New(name, rawURL string) (*Backend, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &Backend{
		Name:    name,
		URL:     u,
		Proxy:   httputil.NewSingleHostReverseProxy(u),
		healthy: true,
	}, nil
}

func (b *Backend) IsHealthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.healthy
}

func (b *Backend) SetHealthy(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.healthy = healthy
}

func (b *Backend) CheckHealth(client *http.Client) bool {
	resp, err := client.Get(b.URL.String() + "/health")

	if err != nil {
		b.SetHealthy(false)
		return false
	}

	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300

	b.SetHealthy(healthy)

	return healthy
}
