# NimbusLB Design Decisions

## 1. Purpose

NimbusLB was designed as a small production-oriented HTTP load balancer that demonstrates practical systems and networking concepts.

The implementation intentionally favors clarity, testability, and predictable behavior over unnecessary complexity.

---

## 2. Why Go?

Go was selected because it provides several features that are particularly useful for networking systems:

* built-in HTTP server support
* efficient concurrency
* lightweight goroutines
* synchronization primitives
* strong networking libraries
* built-in testing and benchmarking
* straightforward deployment

Go also allows NimbusLB to be compiled into a small standalone binary, making it suitable for containerized deployment.

---

## 3. Why Round Robin?

NimbusLB initially uses round-robin scheduling.

For two healthy backends:

```text
Request 1 → Backend A
Request 2 → Backend B
Request 3 → Backend A
Request 4 → Backend B
```

Round robin was selected because it provides:

* deterministic behavior
* simple implementation
* predictable distribution
* low scheduling overhead
* easy testing

It also provides a useful baseline for future scheduling algorithms.

---

## 4. Health-Aware Scheduling

Simple round robin is insufficient when backend instances can fail.

NimbusLB therefore combines round-robin scheduling with backend health state.

Conceptually:

```text
Find next backend
       |
       v
Is it healthy?
   /       \
 Yes        No
  |          |
  v          v
Select     Continue
backend    searching
```

This prevents requests from being intentionally routed to known unhealthy instances.

---

## 5. Active Health Checks

NimbusLB performs periodic active health checks.

The health checker runs independently from request handling.

This separation is intentional:

```text
Request Path:

Client → Load Balancer → Backend


Health Path:

Health Checker → Backend
```

The request path therefore does not need to perform a health check for every request.

This reduces request latency while still allowing the system to detect failures periodically.

---

## 6. In-Memory State

Backend health and scheduling state are currently stored in memory.

Advantages:

* low latency
* simple implementation
* no external dependency
* easy local deployment

Trade-off:

If multiple NimbusLB instances were deployed, their state would not automatically be shared.

A future distributed version could introduce shared service discovery or configuration state.

---

## 7. Reverse Proxy Implementation

NimbusLB uses Go's `httputil.ReverseProxy`.

The project does not implement HTTP proxying from scratch.

This decision keeps responsibilities separated.

NimbusLB owns:

```text
Backend selection
Health awareness
Failure handling
Metrics
Observability
```

The reverse proxy owns:

```text
HTTP forwarding
Backend communication
Response forwarding
```

This reduces implementation complexity and allows the project to focus on load-balancing behavior.

---

## 8. Configuration Outside the Code

Backend addresses and server settings are loaded from YAML configuration.

For example:

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

Keeping configuration outside the application code allows the same binary to be deployed in different environments.

Docker uses a separate configuration because backend services are addressed using Docker service names.

---

## 9. Request IDs

NimbusLB generates a request ID when one is not supplied by the client.

Example:

```text
X-Request-ID: req-42
```

If the client already provides an ID, NimbusLB preserves it.

This makes it possible to correlate:

```text
Client Request
      |
      v
NimbusLB Logs
      |
      v
Backend Request
```

Request IDs become increasingly useful when the system grows into a distributed architecture.

---

## 10. Metrics

NimbusLB exposes operational metrics through:

```text
/metrics
```

The current implementation tracks request and backend activity.

Metrics were deliberately implemented as an internal component rather than coupling the entire application to a specific external monitoring system.

This leaves room for future Prometheus-compatible metrics without requiring major changes to the routing architecture.

---

## 11. Graceful Shutdown

NimbusLB uses Go's graceful HTTP server shutdown functionality.

Instead of immediately terminating the process, shutdown follows this sequence:

```text
SIGTERM / Interrupt
        |
        v
Stop accepting new requests
        |
        v
Allow active requests to finish
        |
        v
Shutdown HTTP server
        |
        v
Exit
```

A timeout prevents shutdown from hanging indefinitely.

This behavior is particularly important in containerized environments.

---

## 12. Failure Semantics

NimbusLB distinguishes between backend failures and load-balancer failures.

If one backend fails:

```text
Backend A → healthy
Backend B → unhealthy

Traffic → Backend A
```

If all backends fail:

```text
Backend A → unhealthy
Backend B → unhealthy

Traffic → 503 Service Unavailable
```

The load balancer process itself remains running.

This is an important design property because backend failures should not automatically become load-balancer process failures.

---

## 13. Testing Strategy

NimbusLB uses multiple testing layers.

### Unit Testing

Tests individual components and algorithms.

Examples:

* round-robin selection
* unhealthy backend skipping
* no-backend behavior
* configuration validation

### Integration Testing

Tests multiple components working together.

Examples:

* HTTP routing
* backend failover

### Benchmarking

Benchmarks measure:

* backend selection
* concurrent backend selection
* HTTP routing

This provides a performance baseline that can be compared against future changes.

---

## 14. Concurrency Design

HTTP requests may arrive concurrently.

The load balancer therefore protects shared scheduling state using synchronization.

The goal is to keep synchronization limited to the smallest required critical section.

This provides:

```text
Concurrent Requests
        |
        v
Thread-safe Backend Selection
        |
        v
Independent HTTP Proxying
```

The backend-selection benchmarks demonstrate that the scheduling operation has very low allocation overhead.

---

## 15. Docker Design

Docker Compose runs three services:

```text
NimbusLB
Backend A
Backend B
```

The services communicate over an internal Docker network.

NimbusLB connects using:

```text
backend-a:9001
backend-b:9002
```

The host exposes only:

```text
localhost:8080
```

This models a common deployment pattern where the load balancer is the externally accessible component while backend services remain internal.

---

## 16. Trade-offs

### Simplicity vs. Feature Set

NimbusLB intentionally avoids implementing advanced features prematurely.

For example, weighted routing and distributed service discovery are not required to demonstrate the fundamental concepts of load balancing and fault tolerance.

This keeps the core system understandable.

### In-Memory State vs. Distributed State

In-memory state provides speed and simplicity.

However, it limits NimbusLB to a simpler single-instance architecture.

A distributed version would require additional mechanisms for:

* state synchronization
* service discovery
* configuration distribution
* failure detection

### Active Health Checks vs. Passive Detection

Active health checks generate periodic network traffic.

The benefit is that backend failures can be detected before a client request reaches the failed backend.

---

## 17. Future Extensions

Potential future improvements include:

### Load Balancing

* weighted round robin
* least connections
* latency-aware routing

### Reliability

* circuit breakers
* retries
* backoff strategies
* connection limits

### Infrastructure

* dynamic backend registration
* service discovery
* distributed configuration
* multiple load-balancer instances

### Security

* TLS termination
* authentication
* authorization
* request filtering

### Observability

* Prometheus metrics
* distributed tracing
* structured JSON logging
* latency histograms

---

## 18. Engineering Principles

The project follows several principles:

### Separation of Concerns

Routing, health checking, configuration, proxying, middleware, and metrics are separate components.

### Testability

Core functionality is implemented so that it can be tested independently.

### Explicit Failure Handling

Backend failures are treated as expected system events.

### Observable Behavior

Request IDs, logs, metrics, and health endpoints provide visibility into system behavior.

### Incremental Development

The project was developed in stages, adding functionality progressively:

```text
Basic HTTP Server
       ↓
Reverse Proxy
       ↓
Load Balancing
       ↓
Health Checks
       ↓
Configuration
       ↓
Observability
       ↓
Testing
       ↓
Benchmarking
       ↓
Containerization
```

---

## 19. Conclusion

NimbusLB demonstrates how a relatively small Go codebase can incorporate several important systems concepts.

The design emphasizes:

* predictable request routing
* health-aware scheduling
* concurrent request handling
* explicit failure behavior
* operational visibility
* graceful lifecycle management
* automated testing
* performance measurement
* containerized deployment

The architecture also leaves clear paths for extending the system into a more advanced distributed load-balancing platform.
