# SkylerRedis — System Design Document

This document describes the **internal architecture, design decisions, and trade-offs** behind SkylerRedis.  
It focuses on **low-level networking, concurrency models, memory management**, and scalability.

---

## 🎯 Design Goals

- High throughput & low latency
- Predictable performance under load
- Minimal locking (shared-nothing preferred)
- Redis-compatible behavior where feasible
- Educational clarity over feature completeness

---

## 🧠 High-Level Architecture

```
Client
  ↓
TCP Listener
  ↓
Connection Dispatcher
  ↓
Shard Router (hash key)
  ↓
Event Loop (1 OS thread / shard)
  ↓
Shard-local KV Engine
```

### Why this design?
- Avoids global locks
- Improves CPU cache locality
- Scales linearly with number of shards

---

## 🌐 Networking & I/O Model

### TCP Server

- Non-blocking sockets
- Manual file descriptor management
- No goroutine-per-connection model

### I/O Multiplexing

- `epoll` on Linux
- `kqueue` on BSD/macOS

### Trade-offs

| Choice | Benefit | Cost |
|------|--------|------|
| Custom event loop | Predictable latency | More complex code |
| No goroutine per conn | High throughput | Less idiomatic Go |

---

## 📡 RESP Parsing

### Design
- Incremental streaming parser
- State machine–based
- Partial read/write safe

### Trade-offs

| Approach | Reason |
|-------|--------|
| Custom parser | Avoids bufio limits |
| sync.Pool buffers | Reduce GC pressure |

---

## 🗄️ Storage Engine

### Multi-Type Storage

Each data type has its **own storage and logic**:
- Strings
- Lists
- Streams
- Sets

**Why not a unified map?**
- Avoids repeated type assertions
- Reduces complex edge-case handling

Trade-off: some commands require scanning multiple maps.

---

## ⏱️ TTL & Expiration

### Strategy
- Lazy expiration on access
- Periodic sampling-based cleanup

### Why not timers?
- Timers per key do not scale
- Heap-based timers add overhead

---

## 🧮 Probabilistic Data Structures

### Bloom Filter
- Membership tests
- False positives accepted

### Count-Min Sketch
- Approximate frequency counting
- Heavy-hitter detection

### Use cases
- Cache admission
- Hot key detection

---

## 🧹 Cache Eviction

### Policy
- Approximate LRU (clock algorithm)
- Random sampling

### Trade-offs

| Policy | Reason |
|------|--------|
| Approx LRU | O(1) eviction |
| Sampling | Avoid full scan |

---

## 🧵 Concurrency Model

### Shared-Nothing Shards

- Each shard owns its data
- No shared mutable state

### Thread-per-Shard

- One OS thread per shard
- `runtime.LockOSThread`
- One event loop per shard

### Why this works well in Go
- Predictable scheduling
- Avoids lock contention
- Better CPU affinity

---

## 🔁 Replication

### Model
- Single Master
- Multiple Replicas

### Simplifications
- No failover
- No cascading replicas

### Replication Flow

```
Replica → PSYNC → Master
Master → RDB snapshot → Replica
Master → Command propagation → Replica
```

---

## 📊 Performance Characteristics

- ≥ 50k ops/sec locally
- Linear scaling with shards
- Stable p99 latency

---

## ⚖️ Key Trade-offs Summary

| Decision | Why | Cost |
|--------|----|-----|
| Custom epoll loop | Control & speed | More code |
| Thread-per-shard | Scalability | Harder debugging |
| No global locks | Performance | Architectural complexity |

---

## 🧠 Learning Outcomes

SkylerRedis is designed to **understand systems**, not just use them:
- How Redis handles scale
- Why certain algorithms are approximate
- Where performance really comes from

---

> *Good systems are not built from abstractions alone — they are built from trade-offs.*

