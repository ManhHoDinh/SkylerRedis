# SkylerRedis — Redis-compatible Database (Low-level Learning Project, Go)

## 📌 Mục tiêu dự án
SkylerRedis là một dự án học tập nhằm **xây dựng lại Redis từ mức low-level**, tập trung vào:
- Networking & I/O multiplexing
- Event-driven architecture
- In-memory data structures
- Concurrency & scalability
- Cache eviction & probabilistic algorithms

Mục tiêu **không phải clone Redis 100%**, mà là **hiểu rõ tại sao Redis được thiết kế như vậy**.

---

## 🧠 Kiến thức & chủ đề chính (Triển khai bằng Go)

> Ngôn ngữ chính: **Go (Golang)** — tập trung vào hiệu năng, kiểm soát low-level I/O và concurrency.

### 1. Networking & Event Loop (Go)
- TCP server non-blocking với `net` + `x/sys/unix`
- I/O multiplexing: `epoll` (Linux), `kqueue` (BSD/macOS)
- Custom event loop (không phụ thuộc hoàn toàn vào goroutine scheduler)
- Read / write buffer, backpressure

Go stack gợi ý:
- `golang.org/x/sys/unix` (epoll/kqueue)
- `syscall.RawConn` để lấy FD
- Tránh goroutine-per-connection ở hot path

> Output: TCP server xử lý được hàng chục nghìn connection đồng thời.

---

### 2. Redis Serialization Protocol (RESP) — Go
- RESP2 format
- Incremental parsing (partial read)
- Streaming parser (state machine)
- Tránh `bufio.Scanner` (limit token)

Gợi ý:
- Tự viết parser dùng `[]byte` + index
- Dùng `sync.Pool` cho buffer

> Output: Tương thích `redis-cli`.

---

### 3. Core Key-Value Engine — Go
- `GET`, `SET`, `DEL`
- TTL
- Lazy expiration
- Sampling-based auto deletion

Triển khai:
- `map[string]*Entry`
- TTL lưu `expireAt int64` (UnixNano)
- Không dùng `time.AfterFunc` per key

> Lưu ý: **Không dùng timer per key**.

---

### 4. Set Data Structure
Implement các lệnh:
- `SADD`, `SREM`
- `SCARD`
- `SMEMBERS`
- `SISMEMBER`
- `SRAND`, `SPOP`

Chiến lược encoding:
- Small set → array
- Large set → hash table

---

### 5. Probabilistic Data Structures

#### Bloom Filter
- Multiple hash functions
- False positive trade-off
- Memory efficient membership test

#### Count-Min Sketch
- Frequency estimation
- Approximate counting
- Heavy-hitter detection

---

### 6. Cache Eviction — Go
- Memory limit
- Eviction policies
- Approximate LRU (clock)
- Random sampling keys

Triển khai:
- LRU clock dùng counter
- Sampling `rand.Intn(len(map))`

> Mục tiêu: O(1) eviction, không scan toàn bộ keyspace.

---

### 7. Concurrency Model — Go

#### Shared-nothing Architecture
- Không chia sẻ state giữa shards
- Mỗi shard là 1 struct KV riêng

#### Thread-per-shard Model (Go style)
- 1 OS thread / shard (`runtime.LockOSThread`)
- 1 event loop / shard
- Channel **chỉ dùng để control**, không cho hot path data
- Client connection được gắn cố định vào shard

> Tránh global mutex, tránh map shared.

---

## 🏗️ Kiến trúc tổng quát (Go)

```
Client
  ↓
TCP Listener (main thread)
  ↓
Shard Router (hash key)
  ↓
Event Loop (1 OS thread / shard)
  ↓
Go KV Engine (map + custom DS)
```

---

## 📊 Performance Goals
- ≥ 50,000 ops/sec (local stress test)
- Scale tuyến tính theo số shard
- Low latency (p99 ổn định)

---

## 🧪 Testing & Benchmark
- GET-heavy workload
- SET-heavy workload
- Mixed read/write
- TTL-heavy workload
- Stress test với nhiều connection

---

## 📅 Timeline (Mar 2025 – Nov 2025)

- **Mar – Apr**: TCP server, epoll/kqueue, event loop
- **May**: RESP parser
- **Jun**: Core KV + TTL
- **Jul**: Set data structure
- **Aug**: Bloom Filter & Count-Min Sketch
- **Sep**: Cache eviction & LRU
- **Oct**: Thread-per-shard, sharding
- **Nov**: Benchmark, docs, optimization

---

## 🚀 Kết quả mong đợi
- Hiểu sâu Redis internals
- Nâng trình low-level & systems design
- Project đủ mạnh để thảo luận ở level mid/senior backend

---

## 🔗 Tham khảo
- Redis source code
- Codecrafters Redis course
- https://github.com/ManhHoDinh/SkylerRedis

---

> *SkylerRedis không chỉ là database — nó là hành trình học cách xây dựng hệ thống phân tán hiệu nâng cao.*

