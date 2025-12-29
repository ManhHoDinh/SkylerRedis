# SkylerRedis — Redis-compatible Database (Low-level Learning Project, Go)

## 📌 Project Goals
SkylerRedis is a learning project aimed at **rebuilding Redis from the low-level up**, focusing on:
- Networking & I/O multiplexing
- Event-driven architecture
- In-memory data structures
- Concurrency & scalability
- Cache eviction & probabilistic algorithms

The goal is **not to clone Redis 100%**, but rather to **understand why Redis is designed the way it is**.

---

## 🧠 Core Knowledge & Topics (Implemented in Go)

> Primary language: **Go (Golang)** — chosen for performance, control over low-level I/O, and concurrency.

### 1. Networking & Event Loop
- Non-blocking TCP server (`net` + `x/sys/unix`)
- I/O multiplexing: `epoll` (Linux), `kqueue` (BSD/macOS)
- Custom event loop (not fully dependent on goroutine scheduler)
- Read/write buffer, backpressure handling

**Output:** TCP server capable of handling tens of thousands of concurrent connections.

---

### 2. Redis Serialization Protocol (RESP)
- RESP2 format
- Incremental parsing (partial reads)
- Streaming parser (state machine)
- Avoid `bufio.Scanner` (token limits)

**Output:** Compatible with `redis-cli`.

---

### 3. Core Key-Value Engine
- Commands: `GET`, `SET`, `DEL`
- TTL support
- Lazy expiration
- Sampling-based auto deletion

**Note:** Do not use per-key timers.

---

### 4. Set Data Structure
Implemented commands:
- `SADD`, `SREM`
- `SCARD`
- `SMEMBERS`
- `SISMEMBER`
- `SRAND`, `SPOP`

Encoding strategy:
- Small sets → array
- Large sets → hash table

---

### 5. Probabilistic Data Structures
- **Bloom Filter**: multiple hash functions, memory-efficient membership test  
- **Count-Min Sketch**: frequency estimation, approximate counting, heavy-hitter detection  

---

### 6. Cache Eviction
- Memory limits
- Eviction policies
- Approximate LRU (clock)
- Random key sampling

**Goal:** O(1) eviction without scanning the entire keyspace.

---

### 7. Concurrency Model
- **Shared-nothing architecture**: no shared state between shards  
- **Thread-per-shard model**:  
  - 1 OS thread per shard (`runtime.LockOSThread`)  
  - 1 event loop per shard  
  - Channels only for control, not hot-path data  
  - Client connections pinned to shards  

---

## 🏗️ Overall Architecture
Client
↓
TCP Listener (main thread)
↓
Shard Router (hash key)
↓
Event Loop (1 OS thread / shard)
↓
Go KV Engine (map + custom DS)

---

## 📊 Performance Goals
- ≥ 50,000 ops/sec (local stress test)
- Linear scaling with number of shards
- Low latency (stable p99)

---

## 🧪 Testing & Benchmark
- GET-heavy workload
- SET-heavy workload
- Mixed read/write workload
- TTL-heavy workload
- Stress testing with many connections

---

## 📅 Timeline (Mar 2025 – Nov 2025)
- **Mar – Apr**: TCP server, epoll/kqueue, event loop  
- **May**: RESP parser  
- **Jun**: Core KV + TTL  
- **Jul**: Set data structure  
- **Aug**: Bloom Filter & Count-Min Sketch  
- **Sep**: Cache eviction & LRU  
- **Oct**: Thread-per-shard, sharding  
- **Nov**: Benchmarking, documentation, optimization  

---

## 🚀 Expected Outcomes
- Deep understanding of Redis internals
- Improved low-level & systems design skills
- A project strong enough to discuss at mid/senior backend level

---

## 🔗 References
- Redis source code  
- Codecrafters Redis course  
- [SkylerRedis GitHub Repository](https://github.com/ManhHoDinh/SkylerRedis)

---

> *SkylerRedis is not just a database — it’s a journey into building advanced distributed systems from the ground up.*


