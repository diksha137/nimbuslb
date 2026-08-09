package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	requestsTotal    uint64
	requestsSuccess  uint64
	requestsFailed   uint64
	backendRequestsA uint64
	backendRequestsB uint64
}

func New() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncRequests() {
	atomic.AddUint64(&m.requestsTotal, 1)
}

func (m *Metrics) IncSuccess() {
	atomic.AddUint64(&m.requestsSuccess, 1)
}

func (m *Metrics) IncFailed() {
	atomic.AddUint64(&m.requestsFailed, 1)
}

func (m *Metrics) IncBackendRequest(name string) {
	switch name {
	case "Backend A":
		atomic.AddUint64(&m.backendRequestsA, 1)

	case "Backend B":
		atomic.AddUint64(&m.backendRequestsB, 1)
	}
}

func (m *Metrics) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Content-Type",
		"text/plain; version=0.0.4",
	)

	fmt.Fprintf(
		w,
		"nimbuslb_requests_total %d\n",
		atomic.LoadUint64(&m.requestsTotal),
	)

	fmt.Fprintf(
		w,
		"nimbuslb_requests_success_total %d\n",
		atomic.LoadUint64(&m.requestsSuccess),
	)

	fmt.Fprintf(
		w,
		"nimbuslb_requests_failed_total %d\n",
		atomic.LoadUint64(&m.requestsFailed),
	)

	fmt.Fprintf(
		w,
		"nimbuslb_backend_requests_total{backend=\"Backend A\"} %d\n",
		atomic.LoadUint64(&m.backendRequestsA),
	)

	fmt.Fprintf(
		w,
		"nimbuslb_backend_requests_total{backend=\"Backend B\"} %d\n",
		atomic.LoadUint64(&m.backendRequestsB),
	)
}
