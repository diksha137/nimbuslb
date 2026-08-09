# NimbusLB Architecture

## 1. Overview

NimbusLB is a lightweight HTTP load balancer implemented in Go.

The system receives HTTP requests from clients, selects a healthy backend using a round-robin scheduling algorithm, and forwards the request through an HTTP reverse proxy.

At the same time, a background health-checking component periodically evaluates backend availability.

The system is designed around four primary concerns:

1. Request routing
2. Backend health management
3. Observability
4. Reliable server lifecycle management

---

## 2. High-Level Architecture

```text
                         Client
                           |
                           | HTTP Request
                           v
                +-----------------------+
                |       HTTP Server     |
                +-----------+-----------+
                            |
                            v
                +-----------------------+
                |     Middleware        |
                |                       |
                |  Request ID           |
                |  Request Logging      |
                +-----------+-----------+
                            |
                            v
                +-----------------------+
                |     Load Balancer     |
                |                       |
                |   Round Robin         |
                |   Health Filtering    |
                +-----------+-----------+
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
                  +---------+---------+
                            |
                     Health Checker
```

---

## 3. Request Lifecycle

A request follows this sequence:

```text
Client
  |
  v
HTTP Server
  |
  v
Request ID Middleware
  |
  v
Logging Middleware
  |
  v
Load Balancer
  |
  +----> Select next healthy backend
  |
  v
Reverse Proxy
  |
  v
Backend Server
  |
  v
HTTP Response
  |
  v
Client
```

### Step 1: Request arrives

The HTTP server accepts the incoming connection.

### Step 2: Request ID

The request ID middleware checks whether the client supplied an `X-Request-ID`.

If not, NimbusLB generates a new request ID.

Example:

```text
X-Request-ID: req-42
```

This identifier is returned to the client and included in server-side logging.

### Step 3: Logging

The logging middleware records information such as:

* request ID
* HTTP method
* request path
* response status
* request duration

This provides basic request tracing and operational visibility.

### Step 4: Backend Selection

The load balancer selects the next healthy backend.

For two healthy backends:

```text
Request 1 → Backend A
Request 2 → Backend B
Request 3 → Backend A
Request 4 → Backend B
```

### Step 5: Reverse Proxy

The selected backend's reverse proxy forwards the HTTP request.

The response is then returned through NimbusLB to the original client.

---

## 4. Backend Health Management

Backend health is maintained independently from request routing.

A background health checker periodically tests each configured backend.

Example:

```text
Backend A → healthy
Backend B → healthy
```

If Backend B fails:

```text
Backend A → healthy
Backend B → unhealthy
```

The balancer then skips Backend B.

When Backend B recovers:

```text
Backend A → healthy
Backend B → healthy
```

it becomes eligible for future requests again.

This separates health monitoring from request handling and allows the routing layer to make decisions based on current backend state.

---

## 5. Round-Robin Scheduling

The initial scheduling algorithm is round robin.

Conceptually:

```text
index = (index + 1) % number_of_backends
```

However, NimbusLB also considers backend health.

Therefore, the effective algorithm is:

```text
Find the next backend in round-robin order
        |
        v
Is backend healthy?
   /          \
 Yes           No
  |             |
  v             |
Select       Continue
backend      searching
```

If all backends are unhealthy, the balancer returns no backend.

The HTTP layer then responds with:

```text
503 Service Unavailable
```

---

## 6. Concurrency

NimbusLB is designed to serve multiple HTTP requests concurrently.

Go's HTTP server creates concurrent request handling automatically.

The load balancer therefore needs synchronization around shared scheduling state.

The backend-selection operation protects its shared state while keeping the critical section small.

This allows concurrent requests to safely update the round-robin position.

Benchmarks were added to measure this behavior under both normal and parallel execution.

---

## 7. Reverse Proxy Layer

NimbusLB uses Go's `httputil.ReverseProxy`.

The proxy is responsible for:

* forwarding HTTP requests
* communicating with backend servers
* receiving backend responses
* returning responses to the client

The load balancer itself therefore focuses primarily on:

```text
Backend selection
+
Health state
+
Failure handling
```

rather than implementing HTTP proxy functionality from scratch.

---

## 8. Configuration

Configuration is loaded from YAML.

The configuration contains:

```text
Server
  └── Port

Health
  └── Check interval

Backends
  ├── Name
  └── URL
```

Example:

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

Configuration validation prevents invalid startup configurations such as:

* missing server port
* no configured backends
* invalid backend configuration

Docker uses a separate configuration containing Docker service names.

---

## 9. Observability

NimbusLB provides basic observability through:

### Request IDs

Each request receives an identifier.

```text
X-Request-ID: req-12
```

### Request Logging

Example:

```text
request_id=req-12
method=GET
path=/
status=200
duration=7.5ms
```

### Metrics

The `/metrics` endpoint exposes operational counters.

These include information about:

* total requests
* successful requests
* failed requests
* requests routed to individual backends

This provides a foundation for future integration with monitoring systems.

---

## 10. Failure Model

NimbusLB treats backend failure as an expected operational condition rather than a process-level failure.

### One backend fails

```text
Backend A → healthy
Backend B → unhealthy

        |
        v

Traffic → Backend A
```

### All backends fail

```text
Backend A → unhealthy
Backend B → unhealthy

        |
        v

503 Service Unavailable
```

The load balancer process itself remains running.

This separation between backend failure and load-balancer failure is an important reliability property.

---

## 11. Graceful Shutdown

The server listens for operating-system termination signals.

When shutdown is requested:

```text
Termination Signal
       |
       v
Stop accepting new work
       |
       v
Allow active requests to finish
       |
       v
Shutdown HTTP server
       |
       v
Process exits
```

A timeout prevents shutdown from waiting indefinitely.

This is particularly important when NimbusLB runs inside containers or orchestration systems.

---

## 12. Testing Strategy

NimbusLB uses several levels of testing.

### Unit Tests

Used for individual components such as:

* load balancing
* configuration validation
* proxy behavior

### Integration Tests

Used to verify complete HTTP routing behavior.

Examples include:

```text
HTTP routing
Backend failover
```

### Benchmarks

Used to measure:

* backend selection performance
* concurrent backend selection
* HTTP routing overhead

This layered testing strategy makes it possible to detect both functional regressions and performance regressions.

---

## 13. Container Architecture

NimbusLB can be deployed using Docker Compose.

```text
                    Docker Network
                         |
       +-----------------+-----------------+
       |                 |                 |
       v                 v                 v
+-------------+   +-------------+   +-------------+
|   NimbusLB  |   |  Backend A  |   |  Backend B  |
|    :8080    |   |    :9001    |   |    :9002    |
+-------------+   +-------------+   +-------------+
```

NimbusLB communicates with the backend containers using Docker service discovery:

```text
http://backend-a:9001
http://backend-b:9002
```

Only the load balancer needs to be exposed to the host.

---

## 14. Design Trade-offs

### Round Robin vs. Advanced Algorithms

Round robin was selected as the initial scheduling algorithm because it is:

* simple
* deterministic
* easy to test
* easy to benchmark

More advanced strategies can be added later.

### Active Health Checks

Active health checks introduce additional network traffic but provide faster detection of backend failures.

### In-Memory State

NimbusLB currently maintains routing and health state in memory.

This keeps the implementation simple and fast, but means the system is currently designed as a single load-balancer instance rather than a distributed control plane.

---

## 15. Future Architecture

Future versions could evolve toward:

```text
                    Clients
                       |
                       v
              +----------------+
              | Load Balancers |
              +-------+--------+
                      |
          +-----------+-----------+
          |           |           |
          v           v           v
      Backend A   Backend B   Backend C
```

Potential additions include:

* multiple load-balancer instances
* distributed configuration
* service discovery
* circuit breakers
* retries
* weighted routing
* least-connections routing
* TLS termination
* distributed tracing
* Prometheus-compatible metrics
* dynamic backend registration

These features would introduce additional distributed-systems challenges such as state synchronization, consistency, leader election, and failure detection.

---

## 16. Summary

NimbusLB is intentionally designed as a small but complete networking system.

Its architecture separates:

```text
Configuration
     |
     v
HTTP Server
     |
     +--> Middleware
     |
     +--> Load Balancer
     |
     +--> Reverse Proxy
     |
     +--> Health Checker
     |
     +--> Metrics
```

This separation makes the system easier to test, reason about, and extend.

The project demonstrates how fundamental systems concepts can be combined to build a practical fault-aware HTTP service in Go.
