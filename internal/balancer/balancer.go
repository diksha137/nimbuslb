package balancer

import (
	"sync"

	"github.com/diksha137/nimbuslb/internal/backend"
)

type Balancer struct {
	backends []*backend.Backend
	current  int
	mutex    sync.Mutex
}

func New(backends []*backend.Backend) *Balancer {
	return &Balancer{
		backends: backends,
	}
}

func (b *Balancer) NextBackend() *backend.Backend {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.backends) == 0 {
		return nil
	}

	for i := 0; i < len(b.backends); i++ {

		selected := b.backends[b.current]

		b.current = (b.current + 1) % len(b.backends)

		if selected.IsHealthy() {
			return selected
		}
	}

	return nil
}
