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

	selected := b.backends[b.current]

	b.current = (b.current + 1) % len(b.backends)

	return selected
}
