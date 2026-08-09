# NimbusLB

A production-oriented HTTP load balancer written in Go.

NimbusLB distributes HTTP traffic across multiple backend servers using round-robin scheduling, continuously monitors backend health, automatically avoids unhealthy instances, and exposes operational metrics.

The project was built to explore practical systems programming concepts including concurrency, networking, fault tolerance, reverse proxies, health checking, graceful shutdown, observability, testing, benchmarking, and containerized deployment.

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

## Features

* HTTP reverse proxy
* Round-robin load balancing
* Concurrent backend selection
* Periodic backend health checks
* Automatic unhealthy-backend exclusion
* Backend failure handling
* `502 Bad Gateway` handling for failed proxy requests
* `503 Service Unavailable` when no healthy backend exists
* Request ID middleware
* Request logging
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

### 1. Request Routing

Incoming HTTP requests are received by NimbusLB.

The load balancer selects the next healthy backend using round-robin scheduling.

```text
Request 1 → Backend A
Request 2 → Backend B
Request 3 → Backend A
Request 4 → Backend B
```

Unhealthy backends are skipped automatically.

---

### 2. Health Checking

NimbusLB periodically checks configured backend servers.

For example:

```text
Backend A → healthy
Backend B → healthy
```

If Backend B becomes unavailable:

```text
Backend A → healthy
Backend B → unhealthy
```

Traffic is then routed only to Backend A.

When Backend B becomes healthy again, it can rejoin the rotation.

---

### 3. Reverse Proxy

NimbusLB uses Go's HTTP reverse-proxy functionality to forward incoming requests to the selected backend.

The client communicates with NimbusLB rather than directly accessing the backend servers.

```text
Client
  |
  v
NimbusLB
  |
  v
Selected Backend
```

---

### 4. Failure Handling

NimbusLB distinguishes between different failure conditions.

If a backend cannot successfully handle a proxied request, the request fails rather than allowing the load balancer process to crash.

If no healthy backend is available, NimbusLB returns:

```text
503 Service Unavailable
```

This allows the load balancer to remain available even when backend services fail.

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

Docker uses a separate configuration because Docker services communicate using Docker's internal service names:

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

The configuration file can be selected using:

```text
NIMBUSLB_CONFIG
```

If the environment variable is not provided, NimbusLB uses:

```text
configs/config.yaml
```

---

## Running Locally

### Start Backend A

```bash
go run ./cmd/backend --port=9001 --name="Backend A"
```

### Start Backend B

```bash
go run ./cmd/backend --port=9002 --name="Backend B"
```

### Start NimbusLB

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

NimbusLB exposes a health endpoint:

```text
GET /health
```

Test it with:

```bash
curl http://localhost:8080/health
```

The endpoint provides visibility into the health state of configured backends.

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

Metrics include request-related and backend-related information collected by the load balancer.

---

## Request IDs and Logging

NimbusLB generates request IDs for incoming requests.

Example:

```text
X-Request-Id: req-1
```

Clients can also provide their own request ID:

```bash
curl -H "X-Request-ID: my-test-123" http://localhost:8080/
```

Example server logging:

```text
request_id=req-2 method=GET path=/ status=200 duration=7.589ms
```

This makes individual requests easier to trace through the system.

---

## Running with Docker

NimbusLB can run as a complete multi-container application using Docker Compose.

Build and start the system:

```bash
docker compose up --build
```

The Docker deployment consists of:

```text
                         localhost:8080
                               |
                               v
                        +-------------+
                        |  NimbusLB    |
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

Test:

```bash
curl http://localhost:8080/
```

Run multiple requests:

```bash
1..10 | ForEach-Object { curl.exe -s http://localhost:8080/ }
```

Stop the system:

```bash
docker compose down
```

Check running containers:

```bash
docker compose ps
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

The integration tests verify HTTP routing and backend failover behavior.

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

Example benchmark results collected during development:

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

These values are development benchmarks rather than production capacity guarantees. Actual performance depends on hardware, operating system, workload, network conditions, and backend behavior.

---

## Concurrency

The load balancer is designed to handle concurrent requests safely.

Backend selection uses synchronization to protect shared state while maintaining low allocation overhead.

Concurrent benchmarks are included to evaluate backend-selection behavior under parallel execution.

Example:

```text
BenchmarkNextBackendParallel
~50 ns/op
0 B/op
0 allocs/op
```

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

This behavior is covered by the project's integration tests.

---

## Graceful Shutdown

NimbusLB uses Go's HTTP server shutdown mechanisms to perform graceful termination.

When the process receives a termination signal, it:

1. Stops accepting new connections.
2. Allows active requests to complete within the configured timeout.
3. Shuts down the HTTP server.
4. Exits cleanly.

The server uses configured timeouts for:

```text
Read timeout
Write timeout
Idle timeout
Shutdown timeout
```

---

## Design Decisions

### Why Go?

Go was selected because it provides:

* lightweight concurrency
* a strong standard library
* excellent HTTP support
* straightforward networking APIs
* built-in testing and benchmarking
* simple deployment as a compiled binary

---

### Why Round Robin?

Round robin provides a simple and deterministic baseline for distributing traffic across backend servers.

It also makes the behavior easy to test and provides a foundation for more advanced scheduling algorithms.

---

### Why Reverse Proxy?

Using Go's reverse-proxy implementation allows NimbusLB to focus on load-balancing logic while relying on a mature HTTP proxy implementation.

---

### Why Health Checks?

A load balancer should not continue sending traffic to unavailable backend instances.

Periodic health checks allow NimbusLB to dynamically remove failed backends from rotation and reintroduce them when they recover.

---

### Why Configuration Files?

Keeping backend addresses, server ports, and health-check intervals outside the application code makes the system easier to deploy in different environments.

The same application binary can therefore run with different configurations.

---

## Reliability

NimbusLB handles several operational failure scenarios:

* unhealthy backend servers
* unavailable backend connections
* no healthy backend servers
* invalid configuration
* concurrent requests
* graceful process termination

The project includes unit tests and integration tests for core routing and failure behavior.

---

## Development Roadmap

The project was developed incrementally.

### Completed

* [x] Basic HTTP backend
* [x] Reverse proxy
* [x] Round-robin load balancing
* [x] Multiple backend support
* [x] Backend health checking
* [x] Unhealthy backend exclusion
* [x] Configuration management
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

### Future Work

Potential improvements include:

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
* more advanced load-balancing algorithms

---

## What This Project Demonstrates

NimbusLB demonstrates practical application of:

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
* service discovery

The goal is not simply to implement a working load balancer, but to explore how a small networking system can be designed to remain observable, testable, configurable, and resilient to component failures.

---

## License

MIT License
