# SkylerRedis Documentation

## 7. Cache Eviction (Approximate LRU)

SkylerRedis implements a cache eviction mechanism to manage memory usage when the number of stored keys exceeds a configurable limit. This is crucial for preventing out-of-memory errors and maintaining performance in scenarios with high data ingress. The policy implemented is an **approximate Least Recently Used (LRU)**, similar to how Redis itself handles memory eviction.

### Configuration: `--maxmemory`

The memory limit is configured via a command-line flag:

```
./redis_clone --maxmemory <number_of_keys>
```

-   **`--maxmemory <N>`**: Sets the maximum number of keys (specifically, key-value pairs in `memory.Store`) that SkylerRedis will store. If this limit is exceeded, the eviction policy is triggered. A value of `0` (default) means no memory limit is enforced.

### LRU Tracking Mechanism

To implement an approximate LRU policy, SkylerRedis tracks the "last access time" for each key using a global LRU clock and a field within each `Entry`.

#### 1. `Entry.LRU` Field

Each `Entry` struct in `internal/memory/storage.go` now includes an `LRU` field:

```go
// internal/memory/storage.go
type Entry struct {
	Value      string
	ExpiryTime time.Time
	LRU        uint64 // Represents approximated last access time for LRU eviction
}
```

This `LRU` field stores a `uint64` value representing the approximated last access time for that specific key.

#### 2. `memory.LruClock` Global Counter

A global counter, `memory.LruClock`, is maintained in `internal/memory/main.go`:

```go
// internal/memory/main.go
var (
    // ...
    LruClock         uint64 // Global LRU clock for approximate LRU eviction
    // ...
)
```

-   **Initialization & Update**:
    -   Whenever a key is accessed (e.g., via `GET`), its `Entry.LRU` field is updated with the current value of `memory.LruClock`.
    -   Whenever a key is created or modified (e.g., via `SET`), its `Entry.LRU` field is initialized with the current `memory.LruClock`.
    -   After each such operation, `memory.LruClock` is incremented. This ensures that recently accessed/modified keys have higher (more recent) `LRU` values.

### Approximate LRU Eviction Algorithm

The core eviction logic resides in the `EvictKeysByLRU()` function in `internal/memory/eviction.go`. This algorithm is triggered when the number of keys in `memory.Store` exceeds the `MaxMemory` limit.

#### How it Works (`EvictKeysByLRU`):

1.  **Trigger Conditions**:
    -   **Periodic**: A background goroutine (started in `app/main.go`) periodically calls `EvictKeysByLRU()` (e.g., every 100ms), alongside the expired key eviction.
    -   **On Write**: After certain commands that modify or add keys (e.g., `SET`), `EvictKeysByLRU()` is called immediately to prevent the memory from growing significantly beyond the limit.

2.  **Memory Limit Check**: The function first checks if `memory.MaxMemory` is set and if `len(memory.Store)` actually exceeds this limit. If not, it returns.

3.  **Target Size**: Eviction doesn't stop exactly at `MaxMemory`. Instead, it aims to reduce the store size to a `targetSize`, which is `MaxMemory` multiplied by `evictionTargetRatio` (e.g., 95% of `MaxMemory`). This prevents the eviction algorithm from being constantly triggered for minor fluctuations.

4.  **Sampling and Eviction Loop**:
    -   While `len(memory.Store)` is greater than `targetSize`:
        -   A small number of keys (e.g., `lruEvictionSample = 10` keys) are randomly sampled from `memory.Store`. Due to Go's randomized map iteration, simply iterating up to the sample size provides a sufficiently random sample.
        -   Among these sampled keys, the key with the *lowest* `Entry.LRU` value is identified. This key is considered the "least recently used" within that sample.
        -   This identified key is then deleted from `memory.Store`.

This iterative sampling and eviction process continues until the `len(memory.Store)` falls below or equals the `targetSize`. This probabilistic approach ensures efficient memory reclamation without the need for a full scan of the keyspace, adhering to the `O(1)` eviction goal as mentioned in the roadmap.
