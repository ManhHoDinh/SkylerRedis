# SkylerRedis Documentation

## 4. Core Key-Value Engine

The heart of SkylerRedis is its in-memory Key-Value (KV) engine. This document details its architecture, focusing on the central data store, thread safety, and the expiration mechanism. This design prioritizes simplicity and performance, drawing inspiration from Redis's core principles.

### Central Data Store

All key-value data is stored in a single, global map located in the `internal/memory` package:

```go
// internal/memory/main.go
var (
    // ...
    Store = make(map[string]Entry)
    // ...
)
```

The `key` is a simple `string`. The `value` is an `Entry` struct, which holds not just the data but also metadata for its lifecycle.

```go
// internal/memory/storage.go
type Entry struct {
	Value      string
	ExpiryTime time.Time
}
```

- **`Value`**: The `string` value associated with the key.
- **`ExpiryTime`**: A `time.Time` object. If this time is the zero value (`time.Time{}`), the key has no expiry. Otherwise, it represents the absolute time at which the key becomes invalid.

### Thread Safety

As a networked server, SkylerRedis must handle concurrent requests safely. The current architecture uses a single event loop, which processes commands sequentially for all clients. However, we also have background tasks, such as the key eviction mechanism, running in a separate goroutine.

To prevent race conditions between the main event loop and background tasks, all access to the central data stores (`Store`, `Sets`, `BloomFilters`, etc.) is protected by a single global mutex:

```go
// internal/memory/main.go
var (
    // ...
    Mu = sync.Mutex{}
    // ...
)
```

Any command handler that reads from or writes to the shared memory must acquire this lock. The `defer` statement is used to guarantee the lock is always released, even if a function returns early.

**Example from `GET` command:**
```go
// internal/command/get.go
func (Get) Handle(...) {
    // ...
	memory.Mu.Lock()
	defer memory.Mu.Unlock()

	entry, ok := memory.Store[key]
    // ... logic to check expiry and return value
}
```

This simple, coarse-grained locking strategy is effective for now. As the system grows, we may explore more fine-grained locking (e.g., sharding the keyspace with a lock per shard) to further improve concurrency, as outlined in the project roadmap.

### Dual Expiration Strategy

A key feature of a Redis-like database is the ability for keys to expire automatically. SkylerRedis employs a dual strategy to manage this, balancing responsiveness with efficient memory reclamation.

#### 1. Passive (Lazy) Expiration

Keys are checked for expiration whenever they are accessed. The `GET` command (and others like it) contains logic to check if a key's `ExpiryTime` has passed.

- **How it works:** When a client requests a key, the server retrieves the `Entry`. It compares the entry's `ExpiryTime` with the current time (`time.Now()`).
- **Action:** If the key is found to be expired, it is deleted from the `Store` on the spot, and the command behaves as if the key never existed (returning `nil`).

This is a highly efficient strategy as it incurs no overhead on keys that are never accessed again.

#### 2. Active, Sampling-based Expiration

Relying only on lazy expiration is insufficient, as keys that are set once and never accessed again would remain in memory forever, causing a memory leak.

To solve this, SkylerRedis runs a background task that actively purges expired keys. This is implemented in `internal/memory/eviction.go`.

- **How it works:** A background goroutine, started in `app/main.go`, runs on a regular interval (e.g., every 100ms).
- **Sampling:** On each tick, the `EvictRandomKeys` function selects a small, random sample of keys (e.g., 20) from the main `Store`.
- **Action:** It checks the `ExpiryTime` for each key in the sample and deletes any that have expired.

This probabilistic approach ensures that memory from expired keys is eventually reclaimed, without the need to perform a costly scan of the entire keyspace. It is the same core strategy that Redis itself uses for active expiration.
