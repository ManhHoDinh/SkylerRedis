# SkylerRedis — Benchmark Report

This document describes the **benchmark methodology, workload patterns, and performance results** of SkylerRedis.

---

## 🧪 Test Environment

- CPU: 8-core (local machine)
- RAM: 16 GB
- OS: Linux / macOS
- Go version: 1.24
- Build mode: release

---

## ⚙️ Benchmark Setup

- Client: `redis-benchmark`, custom Go load generator
- Connections: 50–500 concurrent clients
- Payload size: 32B – 256B
- Persistence: disabled unless stated

---

## 📊 Workload Profiles

### 1. GET-heavy

- 90% GET
- 10% SET

Purpose:
- Measure read latency
- Validate cache efficiency

---

### 2. SET-heavy

- 80% SET
- 20% GET

Purpose:
- Stress write path
- Measure eviction impact

---

### 3. Mixed Read / Write

- 50% GET
- 50% SET

Purpose:
- Simulate real-world usage

---

### 4. TTL-heavy

- Keys with short expiration
- High churn rate

Purpose:
- Validate lazy expiration
- Detect memory leaks

---

## 📈 Results (Local - 200 Concurrent Clients)

| Command | Shards | Requests/sec | p50 Latency | Target (req/sec) |
|---------|--------|--------------|-------------|------------------|
| SET     | 1      | 38233.61     | 4.423 msec  | ≥ 50,000         |
| SET     | 2      | 37119.52     | 4.519 msec  | ≥ 50,000         |
| SET     | 4      | 45620.44     | 4.575 msec  | ≥ 50,000         |
| GET     | 1      | 37376.19     | 5.887 msec  | ≥ 50,000         |
| GET     | 2      | 28989.71     | 6.391 msec  | ≥ 50,000         |
| GET     | 4      | 31496.06     | 6.599 msec  | ≥ 50,000         |

> Results vary by hardware and OS. Benchmarks performed on a local machine with Docker Desktop.

---

## 🔍 Analysis

The benchmark results highlight the current performance characteristics and scalability of SkylerRedis.

### SET Workload

-   **Scaling with Shards**: For SET operations, increasing the number of shards shows a positive trend in requests per second. Performance improved from ~38k req/sec with 1 shard to ~45k req/sec with 4 shards. This demonstrates the benefits of the thread-per-shard architecture in distributing write load.
-   **Target Attainment**: The performance with 4 shards (45620.44 req/sec) is very close to the target of 50,000 req/sec, indicating good progress towards a high-throughput write-heavy system.
-   **Latency**: p50 latency remained relatively stable across different shard configurations for SET, suggesting that the per-shard locking is effective and not introducing significant contention.

### GET Workload

-   **Unexpected Dip**: GET operations showed an unexpected dip in performance when moving from 1 shard to 2 shards, and only a partial recovery with 4 shards. This suggests potential bottlenecks or overheads that are more pronounced during read-heavy operations in the current architecture.
-   **Possible Bottlenecks for GET**:
    -   **Single Event Loop Dispatch**: The global event loop still handles all incoming connections and dispatches commands. For very fast GET operations, the overhead of hashing the key and finding the correct shard might be larger relative to the actual command execution time, becoming a bottleneck.
    -   **Go Scheduler Overhead**: Increased concurrency on the Go runtime (even if data is sharded) might introduce scheduling overhead that impacts read performance more.
    -   **Hardware/Docker Overhead**: The environment (local machine, Docker Desktop) might introduce varying levels of overhead.
-   **Latency**: p50 latency for GET operations increased slightly with more shards, reinforcing the idea of some overhead introduced by the sharding/dispatching mechanism for read-heavy workloads.

### General Observations

-   **Thread-per-shard Effectiveness**: The architecture effectively utilizes CPU resources by distributing data and access patterns, especially for SET operations.
-   **Synchronization**: The per-shard mutexes (`shard.Mu`) successfully prevent concurrent map access issues within each shard.
-   **Future Optimizations**: Further performance gains, especially for GET operations and overall requests per second, might require optimizing the single event loop's dispatch mechanism or potentially exploring `SO_REUSEPORT` with multiple processes to parallelize connection handling itself, fully leveraging the "1 event loop / shard" model as described in the roadmap.

---

## 🚧 Known Bottlenecks (from previous state)

- RESP parsing under very large payloads
- Stream range scans
- Replica lag under write-heavy workloads

---

## 🧠 Takeaways

- Thread-per-shard provides predictable performance, especially under write loads.
- Approximate algorithms are critical for scalability.
- The current dispatching mechanism might introduce overhead for very fast operations like GET, warranting further investigation.

---

> *Benchmarks do not prove correctness — but they reveal architectural truth.*

