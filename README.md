# NimbusLB

A production-oriented HTTP load balancer written in Go.

NimbusLB distributes HTTP traffic across multiple backend servers using round-robin scheduling, continuously monitors backend health, automatically avoids unhealthy instances, and exposes operational metrics.

The project explores practical systems-programming concepts including **concurrency, networking, fault tolerance, reverse proxies, health checking, graceful shutdown, observability, testing, benchmarking, and containerized deployment**.

---

## Architecture

```text
                         HTTP Clients
                              |
                              v
                    +-------------------+
                    |     NimbusLB      |
                    |                   |
                    | HTTP Server       |
                    | Request ID        |
                    | Logging           |
                    | Metrics           |
                    | Health Checking   |
                    | Round Robin LB    |
                    +---------+---------+
                              |
                    +---------+---------+
                    |                   |
                    v                   v
             +-------------+     +-------------+
             |  Backend A  |     |  Backend B  |
             |    :9001    |     |    :9002    |
             +-------------+     +-------------+
                    ^                   ^
                    |                   |
                    +--------+----------+
                             |
                      Health Checker
```

---

## Key Features

* HTTP reverse proxy using Go's `httputil.ReverseProxy`
* Round-robin load balancing
* Thread-safe backend selection
* Periodic backend health checks
* Automatic unhealthy-backend exclusion
* Backend failure handling
* `503 Service Unavailable` when no healthy backend is available
* Request ID middleware
* Structured request logging
* Request and backend metrics
* Graceful HTTP server shutdown
* YAML-based configuration
* Configuration validation
* Unit tests
* Integration tests
* Failure-path tests
* Performance benchmarks
* Concurrent benchmarks
* Docker and Docker Compose support

---

## Project Structure

```text
NimbusLB/
│
├── cmd/
│   ├── backend/
│   │   └── main.go
│   │
│   └── server/
│       └── main.go
│
├── configs/
│   ├── config.yaml
│   └── config.docker.yaml
│
├── docs/
│   ├── architecture.md
│   └── design-decisions.md
│
├── internal/
│   ├── backend/
│   ├── balancer/
│   ├── config/
│   ├── health/
│   ├── metrics/
│   ├── middleware/
│   ├── proxy/
│   └── server/
│
├── tests/
│   └── integration_test.go
│
├── Dockerfile
├── Dockerfile.backend
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## How It Works

### Request Routing

Incoming HTTP requests are received by NimbusLB and passed through middleware before reaching the load-balancing layer.

The load balancer selects the next healthy backend using round-robin scheduling.

```text
Request 1 → Backend A
Request 2 → Backend B
Request 3 → Backend A
Request 4 → Backend B
```

Unhealthy backends are skipped automatically.

---

### Health Checking

NimbusLB periodically checks configured backend servers.

When both backends are healthy:

```text
Backend A → healthy
Backend B → healthy
```

If Backend B becomes unavailable:

```text
Backend A → healthy
Backend B → unhealthy
```

traffic is routed only to Backend A.

When Backend B recovers, it becomes eligible for future requests again.

---

### Reverse Proxy

NimbusLB uses Go's `httputil.ReverseProxy` to forward requests to the selected backend.

The client communicates with NimbusLB rather than directly accessing backend servers.

```text
Client
  |
  v
NimbusLB
  |
  v
Selected Backend
```

This keeps backend selection and health management separate from the HTTP proxy implementation.

---

### Failure Handling

NimbusLB treats backend failures as expected operational conditions.

If a backend becomes unhealthy, it is removed from the routing rotation.

If no healthy backend is available, NimbusLB returns:

```text
503 Service Unavailable
```

The load-balancer process itself remains running.

---

## Configuration

NimbusLB loads its configuration from YAML.

Example local configuration:

```yaml
server:
  port: 8080

health:
  interval_seconds: 5

backends:
  - name: Backend A
    url: http://localhost:9001

  - name: Backend B
    url: http://localhost:9002
```

Docker uses a separate configuration because containers communicate using Docker service names:

```yaml
server:
  port: 8080

health:
  interval_seconds: 5

backends:
  - name: Backend A
    url: http://backend-a:9001

  - name: Backend B
    url: http://backend-b:9002
```

Configuration validation is performed during startup to detect invalid or incomplete configuration.

---

## Running Locally

### 1. Start Backend A

```bash
go run ./cmd/backend --port=9001 --name="Backend A"
```

### 2. Start Backend B

```bash
go run ./cmd/backend --port=9002 --name="Backend B"
```

### 3. Start NimbusLB

```bash
go run ./cmd/server
```

NimbusLB listens on the port configured in `configs/config.yaml`.

Test the load balancer:

```bash
curl http://localhost:8080/
```

Run multiple requests:

```bash
for i in {1..10}; do curl http://localhost:8080/; done
```

On PowerShell:

```powershell
1..10 | ForEach-Object { curl.exe -s http://localhost:8080/ }
```

Expected behavior:

```text
Hello from Backend A
Hello from Backend B
Hello from Backend A
Hello from Backend B
...
```

---

## Health Endpoint

NimbusLB exposes:

```text
GET /health
```

Test it with:

```bash
curl http://localhost:8080/health
```

Example response:

```json
{"status":"healthy"}
```

The endpoint provides a simple operational health signal for the load balancer.

---

## Metrics

NimbusLB exposes operational metrics through:

```text
GET /metrics
```

Test it with:

```bash
curl http://localhost:8080/metrics
```

Example output:

```text
nimbuslb_requests_total 11
nimbuslb_requests_success_total 11
nimbuslb_requests_failed_total 0
nimbuslb_backend_requests_total{backend="Backend A"} 6
nimbuslb_backend_requests_total{backend="Backend B"} 5
```

These counters provide visibility into request volume, successful and failed requests, and backend traffic distribution.

---

## Request IDs and Logging

NimbusLB generates a request ID when one is not provided by the client.

Example:

```text
X-Request-Id: req-1
```

Clients can also provide their own request ID:

```bash
curl -H "X-Request-ID: my-test-123" http://localhost:8080/
```

Example server log:

```text
request_id=req-2 method=GET path=/ status=200 duration=7.589ms
```

Request IDs make individual requests easier to trace through the system.

---

## Running with Docker

NimbusLB can run as a complete multi-container application using Docker Compose.

Build and start the system:

```bash
docker compose up --build
```

The deployment consists of:

```text
                         localhost:8080
                               |
                               v
                        +-------------+
                        |   NimbusLB   |
                        +------+------+ 
                               |
                    +----------+----------+
                    |                     |
                    v                     v
             +-------------+       +-------------+
             |  backend-a  |       |  backend-b  |
             |    :9001    |       |    :9002    |
             +-------------+       +-------------+
```

Only NimbusLB is exposed to the host.

Test the deployment:

```bash
curl http://localhost:8080/
```

Run multiple requests on PowerShell:

```powershell
1..10 | ForEach-Object { curl.exe -s http://localhost:8080/ }
```

Check running containers:

```bash
docker compose ps
```

Stop the system:

```bash
docker compose down
```

---

## Testing

Run the complete test suite:

```bash
go test ./... -count=1
```

Run integration tests:

```bash
go test ./tests -v
```

The integration tests verify complete HTTP behavior including:

* request routing
* round-robin distribution
* backend failover

Example:

```text
=== RUN   TestLoadBalancerHTTPRouting
--- PASS: TestLoadBalancerHTTPRouting

=== RUN   TestLoadBalancerFailover
--- PASS: TestLoadBalancerFailover

PASS
```

---

## Benchmarks

NimbusLB includes benchmarks for backend selection and HTTP routing.

Run all benchmarks:

```bash
go test ./internal/balancer -bench=Benchmark -benchmem -run=^$
```

A development benchmark on the project hardware produced approximately:

```text
BenchmarkNextBackend
~25 ns/op
0 B/op
0 allocs/op
```

Concurrent backend selection:

```text
BenchmarkNextBackendParallel
~50 ns/op
0 B/op
0 allocs/op
```

HTTP routing:

```text
BenchmarkHTTPRouting
~98.7 µs/op
~44.4 KB/op
77 allocs/op
```

These numbers are development measurements rather than production capacity guarantees. Actual performance depends on hardware, operating system, workload, network conditions, and backend behavior.

---

## Concurrency

NimbusLB is designed to handle concurrent HTTP requests safely.

Backend selection protects shared scheduling state using synchronization while keeping the critical section small.

The project includes concurrent benchmarks to evaluate backend-selection behavior under parallel execution.

The benchmark results demonstrate that backend selection performs with:

```text
0 allocations/op
```

under the tested workload.

---

## Failure Handling

NimbusLB is designed to continue operating when individual backend servers fail.

For example:

```text
             Backend A
                |
             healthy
                |
                v
Client → NimbusLB
                ^
                |
             Backend B
             unhealthy
```

Requests are automatically routed away from unhealthy backends.

If all configured backends become unhealthy:

```text
Client
  |
  v
NimbusLB
  |
  X
No healthy backend
  |
  v
503 Service Unavailable
```

This behavior is covered by the integration test suite.

---

## Graceful Shutdown

NimbusLB uses Go's HTTP server shutdown mechanisms to perform graceful termination.

When the process receives a termination signal, it:

1. Stops accepting new work.
2. Allows active requests to complete within the configured timeout.
3. Shuts down the HTTP server.
4. Exits cleanly.

The server also configures:

```text
Read timeout
Write timeout
Idle timeout
Shutdown timeout
```

This makes the service better suited to containerized environments.

---

## Design Decisions

Detailed architectural decisions are documented in:

```text
docs/architecture.md
docs/design-decisions.md
```

Important design choices include:

### Why Go?

Go provides:

* lightweight concurrency
* strong standard-library networking support
* excellent HTTP support
* synchronization primitives
* built-in testing and benchmarking
* straightforward deployment as a compiled binary

### Why Round Robin?

Round robin provides a simple and deterministic baseline for distributing traffic across backend servers.

It is easy to test, has low scheduling overhead, and provides a foundation for more advanced algorithms.

### Why Reverse Proxy?

Go's reverse-proxy implementation allows NimbusLB to focus on:

```text
Backend selection
Health awareness
Failure handling
Observability
```

while relying on the standard library for HTTP forwarding.

### Why Health Checks?

A load balancer should avoid intentionally sending traffic to unavailable backend instances.

Periodic health checks allow NimbusLB to remove failed backends from rotation and reintroduce them after recovery.

### Why Configuration Files?

Keeping backend addresses, server ports, and health-check intervals outside application code makes the system easier to deploy in different environments.

---

## Reliability

NimbusLB handles several operational failure scenarios:

* unhealthy backend servers
* unavailable backend connections
* no healthy backend servers
* invalid configuration
* concurrent requests
* graceful process termination

The project includes unit, integration, and failure-path tests for core behavior.

---

## Development Roadmap

NimbusLB was developed incrementally.

### Completed

* [x] Basic HTTP backend
* [x] Reverse proxy
* [x] Round-robin load balancing
* [x] Multiple backend support
* [x] Backend health checking
* [x] Unhealthy backend exclusion
* [x] Configuration management
* [x] Configuration validation
* [x] Request IDs
* [x] Request logging
* [x] Metrics endpoint
* [x] Graceful shutdown
* [x] Unit tests
* [x] Integration tests
* [x] Failure-path testing
* [x] Performance benchmarks
* [x] Concurrent benchmarks
* [x] Docker deployment
* [x] Docker Compose deployment
* [x] Architecture documentation
* [x] Design decision documentation

### Potential Future Work

* weighted load balancing
* least-connections scheduling
* dynamic backend registration
* circuit breakers
* retry policies
* connection pooling
* TLS termination
* richer Prometheus-compatible metrics
* distributed tracing
* dynamic configuration reloads
* authentication and authorization
* additional load-balancing algorithms

---

## What This Project Demonstrates

NimbusLB combines several practical systems and networking concepts:

* HTTP networking
* reverse proxies
* concurrent programming
* synchronization
* health monitoring
* fault tolerance
* failure handling
* configuration management
* observability
* graceful shutdown
* unit testing
* integration testing
* benchmarking
* containerization
* Docker service networking

The goal is not simply to implement a working load balancer, but to explore how a small networking system can be designed to remain **observable, testable, configurable, and resilient to component failures**.

---

## License

MIT License
