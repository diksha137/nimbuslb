package health

import (
	"log"
	"net/http"
	"time"

	"github.com/diksha137/nimbuslb/internal/backend"
)

type Checker struct {
	backends []*backend.Backend
	interval time.Duration
	client   *http.Client
}

func NewChecker(
	backends []*backend.Backend,
	interval time.Duration,
) *Checker {
	return &Checker{
		backends: backends,
		interval: interval,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *Checker) Start() {
	go func() {

		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			c.checkAll()

			<-ticker.C
		}
	}()
}

func (c *Checker) checkAll() {

	for _, b := range c.backends {

		healthy := b.CheckHealth(c.client)

		if healthy {
			log.Printf(
				"Health check: %s is healthy",
				b.Name,
			)
		} else {
			log.Printf(
				"Health check: %s is UNHEALTHY",
				b.Name,
			)
		}
	}
}
